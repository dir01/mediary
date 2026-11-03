package torrent

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	anacrolixTorrent "github.com/anacrolix/torrent"
	"github.com/dir01/mediary/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func New(dataDir string, logger *slog.Logger, isDebug bool) (*Downloader, error) {
	return newDownloader(dataDir, logger, isDebug, otel.GetMeterProvider())
}

func newDownloader(dataDir string, logger *slog.Logger, isDebug bool, meterProvider metric.MeterProvider) (*Downloader, error) {
	cfg := anacrolixTorrent.NewDefaultClientConfig()
	cfg.ListenPort = 0
	cfg.DataDir = dataDir
	cfg.Debug = isDebug
	torrentClient, err := anacrolixTorrent.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	d := &Downloader{
		torrentClient: torrentClient,
		dataDir:       dataDir,
		log:           logger,
		stallTimeout:  defaultStallTimeout,
		pollInterval:  time.Second,
		downloads:     &downloadRegistry{downloads: make(map[*activeDownload]struct{})},
	}
	if err := d.registerMetrics(meterProvider.Meter("github.com/dir01/mediary/downloader/torrent")); err != nil {
		_ = torrentClient.Close()
		return nil, err
	}
	var _ service.Downloader = d
	return d, nil
}

type Downloader struct {
	torrentClient  *anacrolixTorrent.Client
	dataDir        string
	log            *slog.Logger
	bootstrapPeers []*anacrolixTorrent.Client

	// stallTimeout is the maximum time Download waits without new data.
	stallTimeout time.Duration
	pollInterval time.Duration

	// downloads keeps the Download calls that run now, for the metrics callback.
	downloads *downloadRegistry
}

// defaultStallTimeout is the time after which Download stops when no data arrives.
const defaultStallTimeout = 15 * time.Minute

// SetStallTimeout changes the time after which Download stops when no data arrives.
func (td *Downloader) SetStallTimeout(d time.Duration) {
	td.stallTimeout = d
}

// AddBootstrapPeer registers a torrent client to peer with on every magnet request.
// Intended for testing: allows a local seeder to be wired in without needing DHT or trackers.
func (td *Downloader) AddBootstrapPeer(peer *anacrolixTorrent.Client) {
	td.bootstrapPeers = append(td.bootstrapPeers, peer)
}

func (td *Downloader) addMagnet(url string) (*anacrolixTorrent.Torrent, error) {
	torr, err := td.torrentClient.AddMagnet(url)
	if err != nil {
		return nil, err
	}
	for _, peer := range td.bootstrapPeers {
		torr.AddClientPeer(peer)
	}
	return torr, nil
}

func (td *Downloader) AcceptsURL(url string) bool {
	return strings.HasPrefix(url, "magnet:")
}

func (td *Downloader) GetMetadata(ctx context.Context, url string) (*service.Metadata, error) {
	torr, err := td.addMagnet(url)
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
		td.log.Error("datadir is not a directory", slog.String("datadir", td.dataDir), slog.String("url", url))
		return nil, fmt.Errorf("datadir %s is not a directory", td.dataDir)
	}

	torr, err := td.addMagnet(url)
	if err != nil {
		td.log.Debug("failed to add magnet", slog.String("url", url), slog.Any("error", err))
		return nil, err
	}

	select {
	case <-ctx.Done():
		td.log.Debug("context cancelled", slog.String("url", url), slog.Any("error", ctx.Err()))
		return nil, fmt.Errorf("no torrent metadata received from peers: %w", ctx.Err())
	case <-time.After(td.stallTimeout):
		return nil, fmt.Errorf("download stalled: no torrent metadata received from peers in %s", td.stallTimeout)
	case <-torr.GotInfo():
		break
	}

	fpMap := make(map[string]struct{}, len(filepaths))
	for _, fp := range filepaths {
		fpMap[fp] = struct{}{}
	}

	var files []*anacrolixTorrent.File
	for _, tf := range torr.Files() {
		if _, exists := fpMap[tf.DisplayPath()]; !exists {
			td.log.Debug("skipping file", slog.String("filepath", tf.DisplayPath()), slog.String("url", url))
			continue
		}
		files = append(files, tf)
	}
	// Check all requested files before a download starts. Otherwise, the files
	// that exist continue to download after this function returns.
	if len(files) != len(fpMap) {
		return nil, fmt.Errorf("torrent has %d of %d requested files", len(files), len(fpMap))
	}
	active := &activeDownload{torr: torr, files: files}
	td.downloads.add(active)
	defer td.downloads.remove(active)
	for _, tf := range files {
		td.log.Debug("downloading file", slog.String("filepath", tf.DisplayPath()), slog.String("url", url))
		tf.Download()
	}

	// Wait until each requested file is complete. Stop if the context is done or
	// if no new bytes arrive during stallTimeout: a torrent without seeders must
	// not keep the worker for the full download timeout.
	lastCompleted := int64(-1)
	lastProgressAt := time.Now()
	ticker := time.NewTicker(td.pollInterval)
	defer ticker.Stop()
	for {
		p := progressOf(files)
		if p.completeFiles == len(files) {
			break
		}
		if p.completedBytes != lastCompleted {
			lastCompleted = p.completedBytes
			lastProgressAt = time.Now()
		}
		td.log.Debug("downloading files", slog.String("url", url), slog.Int64("downloaded", p.completedBytes), slog.Int64("total", p.totalBytes))

		var failure error
		if ctx.Err() != nil {
			failure = fmt.Errorf("download timed out: %s: %w", p, ctx.Err())
		} else if time.Since(lastProgressAt) >= td.stallTimeout {
			failure = fmt.Errorf("download stalled: no data received in %s: %s", td.stallTimeout, p)
		}
		if failure != nil {
			for _, tf := range files {
				tf.SetPriority(anacrolixTorrent.PiecePriorityNone)
			}
			return nil, failure
		}

		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	}
	td.log.Debug("all files downloaded", slog.String("url", url))

	filepathsMap = make(map[string]string)
	for _, f := range filepaths {
		filepathsMap[f] = path.Join(td.dataDir, torr.Name(), f)
	}
	td.log.Debug("filepaths map", slog.String("url", url), slog.Any("filepathsMap", filepathsMap))

	return filepathsMap, nil
}

type downloadProgress struct {
	completeFiles  int
	totalFiles     int
	completedBytes int64
	totalBytes     int64
}

func progressOf(files []*anacrolixTorrent.File) downloadProgress {
	p := downloadProgress{totalFiles: len(files)}
	for _, tf := range files {
		done := tf.BytesCompleted()
		if done == tf.Length() {
			p.completeFiles++
		}
		p.completedBytes += done
		p.totalBytes += tf.Length()
	}
	return p
}

func (p downloadProgress) String() string {
	return fmt.Sprintf("%d of %d files complete (%s of %s)",
		p.completeFiles, p.totalFiles, formatBytes(p.completedBytes), formatBytes(p.totalBytes))
}

func formatBytes(n int64) string {
	const mb = 1024 * 1024
	if n < mb {
		return fmt.Sprintf("%d KB", n/1024)
	}
	return fmt.Sprintf("%d MB", n/mb)
}
