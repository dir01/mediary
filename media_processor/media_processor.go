package media_processor

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	id3v2 "github.com/bogem/id3v2/v2"
	"github.com/dir01/mediary/service"
	"github.com/samber/oops"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func NewFFMpegMediaProcessor(logger *slog.Logger) (service.MediaProcessor, error) {
	return &FFMpegMediaProcessor{log: logger}, nil
}

type FFMpegMediaProcessor struct {
	log *slog.Logger
}

func (conv *FFMpegMediaProcessor) GetInfo(ctx context.Context, filepath string) (info *service.MediaInfo, err error) {
	_, span := otel.Tracer("github.com/dir01/mediary/media_processor").Start(ctx, "media_processor.GetInfo",
		trace.WithAttributes(attribute.String("filepath", filepath)),
	)
	defer span.End()

	info = &service.MediaInfo{}

	if state, err := os.Stat(filepath); err == nil {
		info.FileLenBytes = state.Size()
	} else {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if duration, err := conv.GetDuration(filepath); err == nil {
		info.Duration = duration
	} else {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}

	span.SetAttributes(
		attribute.Int64("file.bytes", info.FileLenBytes),
		attribute.Float64("media.duration_seconds", info.Duration.Seconds()),
	)
	return info, nil
}

func (conv *FFMpegMediaProcessor) Concatenate(ctx context.Context, filepaths []string, audioCodec string) (string, error) {
	ctx, span := otel.Tracer("github.com/dir01/mediary/media_processor").Start(ctx, "media_processor.Concatenate",
		trace.WithAttributes(
			attribute.Int("files.count", len(filepaths)),
			attribute.String("audio_codec", audioCodec),
		),
	)
	defer span.End()

	if len(filepaths) == 0 {
		return "", fmt.Errorf("filepaths cannot be empty")
	}

	firstExtIndex := strings.LastIndex(filepaths[0], ".")
	if firstExtIndex == -1 {
		return "", fmt.Errorf("first filepath must include extension")
	}

	ext := filepaths[0][firstExtIndex:]

	// The result is always MP3: the episode feed gives all episodes as MP3, and
	// only MP3 gets the ID3 chapters and cover picture of AddChapterTags. Other
	// formats (m4a, m4b) are encoded to MP3, so "copy" is not possible for them.
	allMP3 := true
	for _, fp := range filepaths {
		allMP3 = allMP3 && isMP3(fp)
	}
	if !allMP3 {
		ext = ".mp3"
		audioCodec = "libmp3lame"
	}

	errCtx := oops.With("ext", ext, "audioCodec", audioCodec, "filepaths", filepaths)
	logAttrs := []any{
		slog.String("ext", ext),
		slog.String("audioCodec", audioCodec),
		slog.Any("filepaths", filepaths),
	}

	file, err := os.CreateTemp("", "*"+ext)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", errCtx.Wrapf(err, "failed to create temp file")
	}
	_ = file.Close()
	resultFilepath := file.Name()
	errCtx = errCtx.With("resultFilepath", resultFilepath)
	logAttrs = append(logAttrs, slog.String("resultFilepath", resultFilepath))
	span.SetAttributes(attribute.String("result.filepath", resultFilepath))

	// The concat protocol joins the bytes of the files. This is correct for MP3,
	// because each MP3 frame is independent. It is not correct for MP4 (m4a, m4b):
	// each file has its own index (the moov atom), and ffmpeg reads only the
	// first file. For all formats other than MP3, use the concat demuxer.
	var inputArgs []string
	if allMP3 {
		inputArgs = []string{"-i", "concat:" + strings.Join(filepaths, "|")}
	} else {
		listFilepath, listErr := writeConcatList(filepaths)
		if listErr != nil {
			span.RecordError(listErr)
			span.SetStatus(codes.Error, listErr.Error())
			return "", errCtx.Wrapf(listErr, "failed to write concat list")
		}
		defer func() { _ = os.Remove(listFilepath) }()
		// The concat demuxer does not keep the tags (title, artist). Read them
		// from the first file, which is the second input.
		inputArgs = []string{
			"-f", "concat", "-safe", "0", "-i", listFilepath,
			"-i", filepaths[0], "-map_metadata", "1",
		}
	}

	// Copy the cover picture (if there is one) and do not encode it again. The mp3
	// muxer keeps all audio packets in memory until it gets the cover picture. An
	// encoded cover comes only at the end of the input, so memory grows with the
	// input size: 2.6 GB for 1.1 GB of MP3. A copied cover comes at the start.
	args := append([]string{"-y"}, inputArgs...)
	args = append(args,
		"-map", "0:a", "-map", "0:v?", "-c:v", "copy",
		"-acodec", audioCodec,
	)

	// When re-encoding (audioCodec != "copy"), probe the source bitrate and
	// pass it to ffmpeg so the output matches the source quality. Without this,
	// ffmpeg uses its default bitrate (128kbps for MP3) which can inflate the output.
	if audioCodec != "copy" {
		if bitrate, probeErr := conv.getAudioBitrate(filepaths[0]); probeErr == nil {
			args = append(args, "-b:a", bitrate)
			span.SetAttributes(attribute.String("source_bitrate", bitrate))
			logAttrs = append(logAttrs, slog.String("sourceBitrate", bitrate))
		} else {
			conv.log.Warn("failed to probe source bitrate, using ffmpeg default",
				append(logAttrs, slog.Any("error", probeErr))...)
		}
	}

	args = append(args, resultFilepath)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	errCtx = errCtx.With("cmd", cmd.String())
	logAttrs = append(logAttrs, slog.String("cmd", cmd.String()))

	conv.log.Debug("running ffmpeg", logAttrs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", errCtx.With("output", string(output)).Wrapf(err, "failed to run ffmpeg")
	}
	conv.log.Debug("ffmpeg finished successfully", logAttrs...)

	return resultFilepath, nil
}

func isMP3(filepath string) bool {
	return strings.EqualFold(path.Ext(filepath), ".mp3")
}

// writeConcatList writes the input file of the ffmpeg concat demuxer and
// returns its path. See https://ffmpeg.org/ffmpeg-formats.html#concat
func writeConcatList(filepaths []string) (string, error) {
	file, err := os.CreateTemp("", "concat_*.txt")
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	var list strings.Builder
	for _, fp := range filepaths {
		// Each path is in single quotes. A single quote in the path (for example,
		// "Abaddon's Gate") closes the quotes, adds an escaped quote and opens them again.
		list.WriteString("file '" + strings.ReplaceAll(fp, "'", `'\''`) + "'\n")
	}
	if _, err := file.WriteString(list.String()); err != nil {
		_ = os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

// getAudioBitrate probes the audio bitrate of the given file using ffprobe.
// Returns bitrate as a string like "128000" (bits per second).
func (conv *FFMpegMediaProcessor) getAudioBitrate(filepath string) (string, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=bit_rate",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filepath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffprobe failed: %w", err)
	}

	bitrate := strings.TrimSpace(string(out))
	if bitrate == "" || bitrate == "N/A" {
		return "", fmt.Errorf("ffprobe returned no bitrate for %s", filepath)
	}
	return bitrate, nil
}

func (conv *FFMpegMediaProcessor) ExtractCoverArt(filepath string) (coverArtFilePath string, err error) {
	errCtx := oops.With("filepath", filepath)
	coverArtFilePath = filepath + ".jpg"
	errCtx = errCtx.With("coverArtFilePath", coverArtFilePath)

	cmd := exec.Command("ffmpeg", "-i", filepath, "-map", "0:v", "-map", "-0:V", "-c", "copy", "-y", coverArtFilePath)
	errCtx = errCtx.With("cmd", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errCtx.With("output", string(out)).Wrapf(err, "failed to run ffmpeg")
	}

	return coverArtFilePath, nil
}

func (conv *FFMpegMediaProcessor) GetDuration(filepath string) (time.Duration, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filepath,
	)
	errCtx := oops.With("filepath", filepath, "cmd", cmd.String())

	out, err := cmd.CombinedOutput()
	errCtx = errCtx.With("output", string(out))
	if err != nil {
		return 0, errCtx.Wrapf(err, "failed to run ffprobe")
	}

	seconds, err := parseDurationOutput(string(out))
	if err != nil {
		return 0, errCtx.Wrapf(err, "failed to parse duration from ffprobe output")
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

// parseDurationOutput extracts the duration float from ffprobe output.
// ffprobe may emit warning/error lines (e.g. "[bmp @ 0x...] bad magic number")
// before the actual value, so we scan lines in reverse for the last parseable float.
func parseDurationOutput(output string) (float64, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if seconds, err := strconv.ParseFloat(line, 64); err == nil {
			return seconds, nil
		}
	}
	return 0, fmt.Errorf("no valid float found in output: %q", output)
}

func (conv *FFMpegMediaProcessor) AddChapterTags(ctx context.Context, filepath string, chapters []service.Chapter) error {
	_, span := otel.Tracer("github.com/dir01/mediary/media_processor").Start(ctx, "media_processor.AddChapterTags",
		trace.WithAttributes(
			attribute.String("filepath", filepath),
			attribute.Int("chapters.count", len(chapters)),
		),
	)
	defer span.End()

	errCtx := oops.With("filepath", filepath, "chapters", len(chapters))
	logAttrs := []any{slog.String("filepath", filepath), slog.Int("chapters", len(chapters))}

	// ID3 tags are only for MP3. An ID3 tag at the start of an MP4 file makes the
	// file unreadable ("moov atom not found").
	if !isMP3(filepath) {
		return errCtx.Errorf("ID3 chapter tags are only for MP3 files")
	}

	// Parse the existing tag, so that Save keeps its frames: the cover picture (APIC)
	// that Concatenate copies, and the title, artist and album. Without Parse, Save
	// writes a tag with only the chapters.
	tag, err := id3v2.Open(filepath, id3v2.Options{Parse: true})
	if err != nil {
		return errCtx.Wrapf(err, "failed to open file for ID3 tagging")
	}
	defer func() { _ = tag.Close() }()

	tag.SetVersion(4) // ID3v2.4 — supports UTF-8 text encoding natively

	// Replace the chapters of the input (if there are any): ffmpeg can copy them
	// from the first file, and they do not apply to the concatenated file.
	tag.DeleteFrames("CHAP")
	tag.DeleteFrames("CTOC")

	childIDs := make([]string, 0, len(chapters))

	for i, ch := range chapters {
		elementID := fmt.Sprintf("chp%d", i)
		childIDs = append(childIDs, elementID)

		tag.AddChapterFrame(id3v2.ChapterFrame{
			ElementID:   elementID,
			StartTime:   ch.StartTime,
			EndTime:     ch.EndTime,
			StartOffset: id3v2.IgnoredOffset,
			EndOffset:   id3v2.IgnoredOffset,
			Title: &id3v2.TextFrame{
				Encoding: id3v2.EncodingUTF8,
				Text:     ch.Title,
			},
		})
	}

	// Add CTOC (Table of Contents) frame — required by the ID3v2 chapters spec
	// for podcast players to discover and navigate chapters.
	tag.AddFrame("CTOC", ctocFrame{
		ElementID: "toc",
		TopLevel:  true,
		Ordered:   true,
		ChildIDs:  childIDs,
	})

	conv.log.Debug("writing ID3 chapter tags", logAttrs...)
	if err := tag.Save(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return errCtx.Wrapf(err, "failed to save ID3 tags")
	}
	return nil
}

// ctocFrame implements id3v2.Framer for the CTOC (Table of Contents) frame,
// which is not natively supported by the bogem/id3v2 library.
// See http://id3.org/id3v2-chapters-1.0
type ctocFrame struct {
	ElementID string
	TopLevel  bool
	Ordered   bool
	ChildIDs  []string
}

func (f ctocFrame) Size() int {
	size := len(f.ElementID) + 1 // null-terminated element ID
	size += 1                    // flags byte
	size += 1                    // entry count
	for _, id := range f.ChildIDs {
		size += len(id) + 1 // null-terminated child element IDs
	}
	return size
}

func (f ctocFrame) UniqueIdentifier() string {
	return f.ElementID
}

func (f ctocFrame) WriteTo(w io.Writer) (int64, error) {
	var written int64

	// Element ID (null-terminated)
	n, err := io.WriteString(w, f.ElementID+"\x00")
	written += int64(n)
	if err != nil {
		return written, err
	}

	// Flags: bit 1 = top-level, bit 0 = ordered
	var flags byte
	if f.TopLevel {
		flags |= 0x02
	}
	if f.Ordered {
		flags |= 0x01
	}
	n, err = w.Write([]byte{flags})
	written += int64(n)
	if err != nil {
		return written, err
	}

	// Entry count
	n, err = w.Write([]byte{byte(len(f.ChildIDs))})
	written += int64(n)
	if err != nil {
		return written, err
	}

	// Child element IDs (null-terminated)
	for _, id := range f.ChildIDs {
		n, err = io.WriteString(w, id+"\x00")
		written += int64(n)
		if err != nil {
			return written, err
		}
	}

	return written, nil
}
