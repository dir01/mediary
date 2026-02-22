package media_processor

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	id3v2 "github.com/bogem/id3v2/v2"
	"github.com/dir01/mediary/service"
	"github.com/samber/oops"
)

func NewFFMpegMediaProcessor(logger *slog.Logger) (service.MediaProcessor, error) {
	return &FFMpegMediaProcessor{log: logger}, nil
}

type FFMpegMediaProcessor struct {
	log *slog.Logger
}

func (conv *FFMpegMediaProcessor) GetInfo(ctx context.Context, filepath string) (info *service.MediaInfo, err error) {
	info = &service.MediaInfo{}

	if state, err := os.Stat(filepath); err == nil {
		info.FileLenBytes = state.Size()
	} else {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if duration, err := conv.GetDuration(filepath); err == nil {
		info.Duration = duration
	} else {
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}

	return info, nil
}

func (conv *FFMpegMediaProcessor) Concatenate(ctx context.Context, filepaths []string, audioCodec string) (string, error) {
	if len(filepaths) == 0 {
		return "", fmt.Errorf("filepaths cannot be empty")
	}

	firstExtIndex := strings.LastIndex(filepaths[0], ".")
	if firstExtIndex == -1 {
		return "", fmt.Errorf("first filepath must include extension")
	}

	ext := filepaths[0][firstExtIndex:]
	errCtx := oops.With("ext", ext, "audioCodec", audioCodec, "filepaths", filepaths)
	logAttrs := []any{
		slog.String("ext", ext),
		slog.String("audioCodec", audioCodec),
		slog.Any("filepaths", filepaths),
	}

	file, err := os.CreateTemp("", "*"+ext)
	if err != nil {
		return "", errCtx.Wrapf(err, "failed to create temp file")
	}
	resultFilepath := file.Name()
	errCtx = errCtx.With("resultFilepath", resultFilepath)
	logAttrs = append(logAttrs, slog.String("resultFilepath", resultFilepath))

	args := []string{"-y", "-i", "concat:" + strings.Join(filepaths, "|"), "-acodec", audioCodec, resultFilepath}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	errCtx = errCtx.With("cmd", cmd.String())
	logAttrs = append(logAttrs, slog.String("cmd", cmd.String()))

	conv.log.Debug("running ffmpeg", logAttrs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errCtx.With("output", string(output)).Wrapf(err, "failed to run ffmpeg")
	}
	conv.log.Debug("ffmpeg finished successfully", logAttrs...)

	return resultFilepath, nil
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

	secondsStr := strings.TrimSpace(string(out))
	seconds, err := strconv.ParseFloat(secondsStr, 64)
	if err != nil {
		return 0, errCtx.Wrapf(err, "failed to parse duration from ffprobe output")
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

func (conv *FFMpegMediaProcessor) AddChapterTags(_ context.Context, filepath string, chapters []service.Chapter) error {
	errCtx := oops.With("filepath", filepath, "chapters", len(chapters))
	logAttrs := []any{slog.String("filepath", filepath), slog.Int("chapters", len(chapters))}

	tag, err := id3v2.Open(filepath, id3v2.Options{Parse: false})
	if err != nil {
		return errCtx.Wrapf(err, "failed to open file for ID3 tagging")
	}
	defer func() { _ = tag.Close() }()

	tag.SetVersion(4) // ID3v2.4 — supports UTF-8 text encoding natively

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
