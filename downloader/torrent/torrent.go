package torrent

import (
	"context"
	"fmt"
	"github.com/hori-ryota/zaperr"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	anacrolixTorrent "github.com/anacrolix/torrent"
	"github.com/dir01/mediary/service"
	"go.uber.org/zap"
)

func New(dataDir string, logger *zap.Logger, isDebug bool) (*Downloader, error) {
	cfg := anacrolixTorrent.NewDefaultClientConfig()
	cfg.ListenPort = 0
	cfg.DataDir = dataDir
	cfg.Debug = isDebug
	torrentClient, err := anacrolixTorrent.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	d := &Downloader{torrentClient: torrentClient, dataDir: dataDir, log: logger}
	var _ service.Downloader = d
	return d, nil
}

type Downloader struct {
	torrentClient *anacrolixTorrent.Client
	dataDir       string
	log           *zap.Logger
}

func (td *Downloader) AcceptsURL(url string) bool {
	return strings.HasPrefix(url, "magnet:")
}

func (td *Downloader) GetMetadata(ctx context.Context, url string) (*service.Metadata, error) {
	torr, err := td.torrentClient.AddMagnet(url)
	if err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-torr.GotInfo():
		break
	}

	info := torr.Info()
	variants := make([]service.VariantMetadata, len(info.Files))
	for i, f := range info.Files {
		variants[i] = service.VariantMetadata{
			ID:       f.DisplayPath(info),
			LenBytes: &f.Length,
		}
	}
	sort.Slice(variants, func(i, j int) bool {
		return variants[i].ID < variants[j].ID
	})
	metadata := &service.Metadata{
		URL:                   url,
		Name:                  info.Name,
		Variants:              variants,
		AllowMultipleVariants: true,
		DownloaderName:        "torrent",
	}
	return metadata, nil
}

func (td *Downloader) Download(ctx context.Context, url string, filepaths []string) (filepathsMap map[string]string, err error) {
	// if datadir is not a directory, we can't download anything
	if stat, err := os.Stat(td.dataDir); err != nil || stat == nil || !stat.IsDir() {
		td.log.Error("datadir is not a directory", zap.String("datadir", td.dataDir), zap.String("url", url))
		return nil, fmt.Errorf("datadir %s is not a directory", td.dataDir)
	}

	torr, err := td.torrentClient.AddMagnet(url)
	if err != nil {
		td.log.Debug("failed to add magnet", zap.String("url", url), zaperr.ToField(err))
		return nil, err
	}

	select {
	case <-ctx.Done():
		td.log.Debug("context cancelled", zap.String("url", url), zap.Error(ctx.Err()))
		return nil, ctx.Err()
	case <-torr.GotInfo():
		break
	}

	fpMap := make(map[string]struct{}, len(filepaths))
	for _, fp := range filepaths {
		fpMap[fp] = struct{}{}
	}

	var wg sync.WaitGroup
	for _, tf := range torr.Files() {
		tf := tf
		if _, exists := fpMap[tf.DisplayPath()]; !exists {
			td.log.Debug("skipping file", zap.String("filepath", tf.DisplayPath()), zap.String("url", url))
			continue
		} else {
			td.log.Debug("downloading file", zap.String("filepath", tf.DisplayPath()), zap.String("url", url))

			wg.Add(1)
			go func() {
				defer wg.Done()
				tf.Download()
				for {
					select {
					case <-ctx.Done():
						tf.SetPriority(anacrolixTorrent.PiecePriorityNone)
						return
					case <-time.After(1 * time.Second):
						td.log.Debug("downloading file", zap.String("filepath", tf.DisplayPath()), zap.String("url", url), zap.Int64("downloaded", tf.BytesCompleted()), zap.Int64("total", tf.Length()))
						if tf.BytesCompleted() == tf.Length() {
							return
						}
						continue
					}
				}
			}()
		}
	}
	wg.Wait()
	td.log.Debug("all files downloaded", zap.String("url", url))

	filepathsMap = make(map[string]string)
	for _, f := range filepaths {
		filepathsMap[f] = path.Join(td.dataDir, torr.Name(), f)
	}
	td.log.Debug("filepaths map", zap.String("url", url), zap.Any("filepathsMap", filepathsMap))

	return filepathsMap, nil
}
