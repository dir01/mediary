package ytdl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hori-ryota/zaperr"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

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

func New(ytdlDir string, dataDir string, logger *zap.Logger) (*YtdlDownloader, error) {
	d := &YtdlDownloader{dataDir: dataDir, ytdlDir: ytdlDir, log: logger}
	var _ service.Downloader = d
	return d, nil
}

type YtdlDownloader struct {
	// directory where the downloaded files will be stored (tempdir)
	dataDir string
	// directory where the https://github.com/ytdl-org/youtube-dl repo is cloned to
	// we use cloned version because official version is too outdated
	ytdlDir string
	log     *zap.Logger
}

func (y *YtdlDownloader) AcceptsURL(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := y.GetMetadata(ctx, url)
	if err != nil {
		return false
	}
	return true
}

func (y *YtdlDownloader) GetMetadata(ctx context.Context, url string) (*service.Metadata, error) {
	out, err := y.runYTDL(ctx, "--dump-json", url)
	if err != nil {
		return nil, err
	}

	var ytdljson ytdlJSON
	if err = json.Unmarshal(out, &ytdljson); err != nil {
		return nil, err
	}

	variants := []service.VariantMetadata{
		{ID: formatTypeVideo},
		{ID: formatTypeAudioHQ},
		{ID: formatTypeAudioMQ},
		{ID: formatTypeAudioLQ},
	}

	return &service.Metadata{
		URL:                   ytdljson.WebpageUrl,
		Name:                  ytdljson.Title,
		Variants:              variants,
		AllowMultipleVariants: false,
	}, nil
}

func (y *YtdlDownloader) Download(ctx context.Context, url string, filepaths []string) (filepathsMap map[string]string, err error) {
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

	if _, err := y.runYTDL(ctx, args...); err != nil {
		return nil, err
	}

	return map[string]string{ytFormat: destinationPath}, nil
}

func (y *YtdlDownloader) runYTDL(ctx context.Context, args ...string) (out []byte, err error) {
	ytdlPath := path.Join(y.ytdlDir, "bin", "youtube-dl")
	args = append([]string{ytdlPath}, args...)
	cmd := exec.CommandContext(context.Background(), "python3", args...)
	cmd.Env = append(cmd.Env, "PYTHONPATH="+y.ytdlDir, "PATH="+os.Getenv("PATH"))
	out, err = cmd.CombinedOutput()
	if err != nil {
		return nil, zaperr.Wrap(
			err,
			"failed to run youtube-dl",
			zap.Strings("args", args),
			zap.String("combined_output", string(out)),
			zap.Strings("env", cmd.Env),
		)
	}
	return out, nil
}

func (y *YtdlDownloader) findAudioFormat(formats []format) *format {
	audioFormats := make([]format, 0, len(formats))
	for _, f := range formats {
		if f.Acodec != "" && f.Vcodec == "none" {
			audioFormats = append(audioFormats, f)
		}
	}
	if len(audioFormats) == 0 {
		return nil
	}
	bestFormat := audioFormats[0]
	for _, f := range audioFormats {
		if !strings.Contains(bestFormat.Acodec, "m4a") && strings.Contains(f.Acodec, "mp4a") {
			bestFormat = f
			continue
		}
		if strings.Contains(f.Acodec, "m4a") && strings.Contains(bestFormat.Acodec, "m4a") && f.Abr > bestFormat.Abr {
			bestFormat = f
			continue
		}
	}
	return &bestFormat
}
