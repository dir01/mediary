package media_processor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	id3v2 "github.com/bogem/id3v2/v2"
	"github.com/dir01/mediary/service"
	"github.com/hori-ryota/zaperr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func NewFFMpegMediaProcessor(logger *zap.Logger) (service.MediaProcessor, error) {
	return &FFMpegMediaProcessor{log: logger}, nil
}

type FFMpegMediaProcessor struct {
	log *zap.Logger
}

func (conv *FFMpegMediaProcessor) GetInfo(ctx context.Context, filepath string) (info *service.MediaInfo, err error) {
	info = &service.MediaInfo{}

	if state, err := os.Stat(filepath); err == nil {
		info.FileLenBytes = state.Size()
	} else {
		return nil, errors.Wrap(err, "failed to stat file")
	}

	if duration, err := conv.GetDuration(filepath); err == nil {
		info.Duration = duration
	} else {
		return nil, errors.Wrap(err, "failed to get duration")
	}

	return info, nil
}

func (conv *FFMpegMediaProcessor) Concatenate(ctx context.Context, filepaths []string, audioCodec string) (string, error) {
	ext := filepaths[0][strings.LastIndex(filepaths[0], "."):] // FIXME
	zapFields := []zap.Field{
		zap.String("ext", ext),
		zap.String("audioCodec", audioCodec),
		zap.Strings("filepaths", filepaths),
	}

	file, err := os.CreateTemp("", "*"+ext)
	if err != nil {
		return "", zaperr.Wrap(err, "failed to create temp file", zapFields...)
	}
	resultFilepath := file.Name()
	zapFields = append(zapFields, zap.String("resultFilepath", resultFilepath))

	args := []string{"-y", "-i", "concat:" + strings.Join(filepaths, "|"), "-acodec", audioCodec, resultFilepath}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	zapFields = append(zapFields, zap.String("cmd", cmd.String()))

	conv.log.Debug("running ffmpeg", zapFields...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", zaperr.Wrap(err, "failed to run ffmpeg", []zap.Field{zap.String("output", string(output))}...)
	}
	conv.log.Debug("ffmpeg finished successfully", zapFields...)

	return resultFilepath, nil
}

func (conv *FFMpegMediaProcessor) ExtractCoverArt(filepath string) (coverArtFilePath string, err error) {
	zapFields := []zap.Field{zap.String("filepath", filepath)}
	coverArtFilePath = filepath + ".jpg"
	zapFields = append(zapFields, zap.String("coverArtFilePath", coverArtFilePath))

	cmd := exec.Command("ffmpeg", "-i", filepath, "-map", "0:v", "-map", "-0:V", "-c", "copy", "-y", coverArtFilePath)
	zapFields = append(zapFields, zap.String("cmd", cmd.String()))

	out, err := cmd.CombinedOutput()
	zapFields = append(zapFields, zap.String("output", string(out)))
	if err != nil {
		return "", zaperr.Wrap(err, "failed to run ffmpeg", zapFields...)
	}

	return coverArtFilePath, nil
}

func (conv *FFMpegMediaProcessor) GetDuration(filepath string) (time.Duration, error) {
	cmd := exec.Command("ffmpeg", "-v", "quiet", "-stats", "-i", filepath, "-f", "null", "-")
	zapFields := []zap.Field{zap.String("filepath", filepath), zap.String("cmd", cmd.String())}

	out, err := cmd.CombinedOutput()
	zapFields = append(zapFields, zap.String("output", string(out)))
	if err != nil {
		return 0, zaperr.Wrap(err, "failed to run ffmpeg", zapFields...)
	}

	re := regexp.MustCompile(`(\d\d:\d\d:\d\d)`)
	found := re.FindAll(out, -1)
	if len(found) == 0 {
		return 0, zaperr.Wrap(err, "failed to parse duration", zapFields...)
	}
	lastFound := found[len(found)-1]
	zapFields = append(zapFields, zap.String("lastFound", string(lastFound)))
	tsParts := strings.Split(string(lastFound), ":")

	hours, err := strconv.Atoi(tsParts[0])
	if err != nil {
		return 0, zaperr.Wrap(err, "failed to parse hours", zapFields...)
	}
	minutes, err := strconv.Atoi(tsParts[1])
	if err != nil {
		return 0, zaperr.Wrap(err, "failed to parse minutes", zapFields...)
	}
	seconds, err := strconv.Atoi(tsParts[2])
	if err != nil {
		return 0, zaperr.Wrap(err, "failed to parse seconds", zapFields...)
	}

	return time.Duration(seconds+60*minutes+60*60*hours) * time.Second, nil
}

func (conv *FFMpegMediaProcessor) AddChapterTags(_ context.Context, filepath string, chapters []service.Chapter) error {
	zapFields := []zap.Field{zap.String("filepath", filepath), zap.Int("chapters", len(chapters))}

	tag, err := id3v2.Open(filepath, id3v2.Options{Parse: false})
	if err != nil {
		return zaperr.Wrap(err, "failed to open file for ID3 tagging", zapFields...)
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

	conv.log.Debug("writing ID3 chapter tags", zapFields...)
	if err := tag.Save(); err != nil {
		return zaperr.Wrap(err, "failed to save ID3 tags", zapFields...)
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
