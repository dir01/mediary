package media_processor

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	if len(filepaths) == 0 {
		return "", errors.New("filepaths cannot be empty")
	}

	firstExtIndex := strings.LastIndex(filepaths[0], ".")
	if firstExtIndex == -1 {
		return "", errors.New("first filepath must include extension")
	}

	ext := filepaths[0][firstExtIndex:]
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
