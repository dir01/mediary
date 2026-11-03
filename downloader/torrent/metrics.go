package torrent

import (
	"context"
	"fmt"
	"sync"

	anacrolixTorrent "github.com/anacrolix/torrent"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// activeDownload is a Download call that waits for the data of its files.
type activeDownload struct {
	torr  *anacrolixTorrent.Torrent
	files []*anacrolixTorrent.File
}

// downloadRegistry keeps the Download calls that run now. The metrics callback reads it.
type downloadRegistry struct {
	mu        sync.Mutex
	downloads map[*activeDownload]struct{}
}

func (r *downloadRegistry) add(d *activeDownload) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.downloads[d] = struct{}{}
}

func (r *downloadRegistry) remove(d *activeDownload) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.downloads, d)
}

// byTorrent groups the running downloads by torrent. Two Download calls can use
// the same torrent, so each torrent gets the union of the requested files.
func (r *downloadRegistry) byTorrent() map[*anacrolixTorrent.Torrent][]*anacrolixTorrent.File {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make(map[*anacrolixTorrent.Torrent][]*anacrolixTorrent.File)
	seen := make(map[*anacrolixTorrent.File]struct{})
	for d := range r.downloads {
		files := result[d.torr]
		for _, f := range d.files {
			if _, dup := seen[f]; dup {
				continue
			}
			seen[f] = struct{}{}
			files = append(files, f)
		}
		result[d.torr] = files
	}
	return result
}

var (
	stateMetadata    = metric.WithAttributes(attribute.String("torrent.state", "metadata"))
	stateDownloading = metric.WithAttributes(attribute.String("torrent.state", "downloading"))
	stateIdle        = metric.WithAttributes(attribute.String("torrent.state", "idle"))
	peerKindActive   = attribute.String("peer.kind", "active")
	peerKindSeeders  = attribute.String("peer.kind", "seeders")
)

// registerMetrics creates the torrent instruments and one callback that observes all of them.
func (td *Downloader) registerMetrics(meter metric.Meter) error {
	bytesRead, err := meter.Int64ObservableCounter("mediary.torrent.bytes_read",
		metric.WithDescription("Useful data bytes that the torrent client received from peers"),
		metric.WithUnit("By"))
	if err != nil {
		return fmt.Errorf("failed to create mediary.torrent.bytes_read: %w", err)
	}
	bytesWritten, err := meter.Int64ObservableCounter("mediary.torrent.bytes_written",
		metric.WithDescription("Data bytes that the torrent client sent to peers"),
		metric.WithUnit("By"))
	if err != nil {
		return fmt.Errorf("failed to create mediary.torrent.bytes_written: %w", err)
	}
	torrents, err := meter.Int64ObservableGauge("mediary.torrents",
		metric.WithDescription("Torrents in the torrent client, by state"),
		metric.WithUnit("{torrent}"))
	if err != nil {
		return fmt.Errorf("failed to create mediary.torrents: %w", err)
	}
	peers, err := meter.Int64ObservableGauge("mediary.torrent.peers",
		metric.WithDescription("Peers connected over all torrents"),
		metric.WithUnit("{peer}"))
	if err != nil {
		return fmt.Errorf("failed to create mediary.torrent.peers: %w", err)
	}
	dlCompleted, err := meter.Int64ObservableGauge("mediary.download.bytes_completed",
		metric.WithDescription("Completed bytes of the requested files of a running download"),
		metric.WithUnit("By"))
	if err != nil {
		return fmt.Errorf("failed to create mediary.download.bytes_completed: %w", err)
	}
	dlTotal, err := meter.Int64ObservableGauge("mediary.download.bytes_total",
		metric.WithDescription("Total bytes of the requested files of a running download"),
		metric.WithUnit("By"))
	if err != nil {
		return fmt.Errorf("failed to create mediary.download.bytes_total: %w", err)
	}
	dlBytesRead, err := meter.Int64ObservableCounter("mediary.download.bytes_read",
		metric.WithDescription("Useful data bytes that the torrent of a running download received from peers"),
		metric.WithUnit("By"))
	if err != nil {
		return fmt.Errorf("failed to create mediary.download.bytes_read: %w", err)
	}
	dlPeers, err := meter.Int64ObservableGauge("mediary.download.peers",
		metric.WithDescription("Peers connected to the torrent of a running download"),
		metric.WithUnit("{peer}"))
	if err != nil {
		return fmt.Errorf("failed to create mediary.download.peers: %w", err)
	}

	_, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		clientStats := td.torrentClient.Stats()
		o.ObserveInt64(bytesRead, clientStats.BytesReadUsefulData.Int64())
		o.ObserveInt64(bytesWritten, clientStats.BytesWrittenData.Int64())
		o.ObserveInt64(peers, int64(clientStats.ActivePeers), metric.WithAttributes(peerKindActive))
		o.ObserveInt64(peers, int64(clientStats.ConnectedSeeders), metric.WithAttributes(peerKindSeeders))

		running := td.downloads.byTorrent()
		var numMetadata, numDownloading, numIdle int64
		for _, t := range td.torrentClient.Torrents() {
			switch _, isRunning := running[t]; {
			case isRunning:
				numDownloading++
			case t.Info() == nil:
				numMetadata++
			default:
				numIdle++
			}
		}
		o.ObserveInt64(torrents, numMetadata, stateMetadata)
		o.ObserveInt64(torrents, numDownloading, stateDownloading)
		o.ObserveInt64(torrents, numIdle, stateIdle)

		for t, files := range running {
			name := attribute.String("torrent.name", t.Name())
			p := progressOf(files)
			stats := t.Stats()
			o.ObserveInt64(dlCompleted, p.completedBytes, metric.WithAttributes(name))
			o.ObserveInt64(dlTotal, p.totalBytes, metric.WithAttributes(name))
			o.ObserveInt64(dlBytesRead, stats.BytesReadUsefulData.Int64(), metric.WithAttributes(name))
			o.ObserveInt64(dlPeers, int64(stats.ActivePeers), metric.WithAttributes(name, peerKindActive))
			o.ObserveInt64(dlPeers, int64(stats.ConnectedSeeders), metric.WithAttributes(name, peerKindSeeders))
		}
		return nil
	}, bytesRead, bytesWritten, torrents, peers, dlCompleted, dlTotal, dlBytesRead, dlPeers)
	if err != nil {
		return fmt.Errorf("failed to register torrent metrics callback: %w", err)
	}
	return nil
}
