package media_processor

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dir01/mediary/service"
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

	file, err := ioutil.TempFile("", "*"+ext)
	if err != nil {
		return "", fmt.Errorf("failed to get temp file: %w", err)
	}
	resultFilepath := file.Name()

	args := []string{"-y", "-i", "concat:" + strings.Join(filepaths, "|"), "-acodec", audioCodec, resultFilepath}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	conv.log.Debug("runing ffmpeg", zap.Any("cmd", cmd))
	if output, err := cmd.CombinedOutput(); err != nil {
		conv.log.Error(
			"failed to run ffmpeg",
			zap.String("cmd", cmd.String()),
			zap.String("output", string(output)),
			zap.String("err", err.Error()),
		)
		return "", err
	} else {
		conv.log.Debug("ffmpeg finished successfully", zap.String("output", string(output)))
	}

	return resultFilepath, nil
}

func (conv *FFMpegMediaProcessor) ExtractCoverArt(filepath string) (coverArtFilePath string, err error) {
	//fullPath := path.Join(conv.workDir, filepath)
	coverArtFilePath = filepath + ".jpg"
	cmd := exec.Command("ffmpeg", "-i", filepath, "-map", "0:v", "-map", "-0:V", "-c", "copy", "-y", coverArtFilePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("ffmpeg failed to extract cover art: %s", string(out))
		return "", err
	}
	return coverArtFilePath, nil
}

func (conv *FFMpegMediaProcessor) GetDuration(filepath string) (time.Duration, error) {
	cmd := exec.Command("ffmpeg", "-v", "quiet", "-stats", "-i", filepath, "-f", "null", "-")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, errors.Wrapf(err, "ffmpeg failed to get durationd: %s", string(out))
	}
	re := regexp.MustCompile(`(\d\d:\d\d:\d\d)`)
	found := re.FindAll(out, -1)
	if len(found) == 0 {
		return 0, fmt.Errorf("no timestamp found in %s", string(out))
	}
	lastFound := found[len(found)-1]
	tsParts := strings.Split(string(lastFound), ":")

	hours, err := strconv.Atoi(tsParts[0])
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse hours of %s (full output: %s)", lastFound, string(out))
	}
	minutes, err := strconv.Atoi(tsParts[1])
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse minutes of %s (full output: %s)", lastFound, string(out))
	}
	seconds, err := strconv.Atoi(tsParts[2])
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse seconds of %s (full output: %s)", lastFound, string(out))
	}

	return time.Duration(seconds+60*minutes+60*60*hours) * time.Second, nil
}
