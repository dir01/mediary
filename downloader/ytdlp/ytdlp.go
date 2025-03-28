package ytdlp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/hori-ryota/zaperr"

	"github.com/dir01/mediary/service"
	"go.uber.org/zap"
)

var ErrNoAudioFormat = fmt.Errorf("could not find an audio format")

const (
	formatTypeVideo   = "Video (mp4)"
	formatTypeAudioHQ = "Audio (mp3), High Quality"
	formatTypeAudioMQ = "Audio (mp3), Medium Quality"
	formatTypeAudioLQ = "Audio (mp3), Low Quality"
)

func New(dataDir string, logger *zap.Logger) (*YtdlpDownloader, error) {
	d := &YtdlpDownloader{dataDir: dataDir, log: logger}
	var _ service.Downloader = d
	return d, nil
}

type YtdlpDownloader struct {
	// dataDir is a location for temporary storage of downloaded files
	dataDir string
	log     *zap.Logger
}

func (y *YtdlpDownloader) AcceptsURL(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := y.GetMetadata(ctx, url)
	if err != nil {
		y.log.Debug("yt-dlp get metadata", zap.Error(err))

		return false
	}

	return true
}

func (y *YtdlpDownloader) GetMetadata(ctx context.Context, url string) (*service.Metadata, error) {
	out, err := y.runYTDLP(ctx, "--dump-json", url)
	if err != nil {
		return nil, err
	}

	var ytdlpjson ytdlpJSON

	if err = json.Unmarshal(out, &ytdlpjson); err != nil {
		return nil, err
	}

	y.log.Debug("metadata unmarshal success", zap.String("url", url))

	variants := []service.VariantMetadata{
		{ID: formatTypeVideo},
		{ID: formatTypeAudioHQ},
		{ID: formatTypeAudioMQ},
		{ID: formatTypeAudioLQ},
	}

	return &service.Metadata{
		URL:                   ytdlpjson.WebpageUrl,
		Name:                  ytdlpjson.Title,
		Variants:              variants,
		AllowMultipleVariants: false,
		DownloaderName:        "ytdl",
	}, nil
}

func (y *YtdlpDownloader) Download(ctx context.Context, url string, filepaths []string) (filepathsMap map[string]string, err error) {
	if len(filepaths) != 1 {
		return nil, fmt.Errorf("expected 1 filepath, got %d", len(filepaths))
	}

	args := []string{url, "--prefer-ffmpeg"}
	destinationPathBase := path.Join(y.dataDir, uuid.New().String())
	var destinationPath string
	ytFormat := filepaths[0]
	switch ytFormat {
	case formatTypeVideo:
		destinationPath = destinationPathBase + ".mp4"
		args = append(args, "--format", "mp4")
	case formatTypeAudioHQ:
		destinationPath = destinationPathBase + ".mp3"
		args = append(args, "--extract-audio", "--audio-format", "mp3", "--audio-quality", "0")
	case formatTypeAudioMQ:
		destinationPath = destinationPathBase + ".mp3"
		args = append(args, "--extract-audio", "--audio-format", "mp3", "--audio-quality", "5")
	case formatTypeAudioLQ:
		destinationPath = destinationPathBase + ".mp3"
		args = append(args, "--extract-audio", "--audio-format", "mp3", "--audio-quality", "9")
	default:
		return nil, fmt.Errorf("unknown format: %s", ytFormat)
	}
	args = append(args, "--output", destinationPath)

	if _, err := y.runYTDLP(ctx, args...); err != nil {
		return nil, err
	}

	return map[string]string{ytFormat: destinationPath}, nil
}

func (y *YtdlpDownloader) runYTDLP(ctx context.Context, args ...string) (out []byte, err error) {
	y.log.Debug("running yt-dlp", zap.Strings("args", args))

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))

	out, err = cmd.CombinedOutput()
	if err != nil {
		return nil, zaperr.Wrap(
			err,
			"failed to run yt-dlp",
			zap.Strings("args", args),
			zap.String("combined_output", string(out)),
			zap.Strings("env", cmd.Env),
		)
	}

	return out, nil
}
