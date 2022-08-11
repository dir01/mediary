package mediary

import (
	"context"
	"strings"

	"github.com/anacrolix/torrent"
)

func NewTorrentDownloader() (*TorrentDownloader, error) {
	cfg := torrent.NewDefaultClientConfig()
	cfg.ListenPort = 0
	torrentClient, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	d := &TorrentDownloader{torrentClient: torrentClient}
	var _ Downloader = d
	return d, nil
}

type TorrentDownloader struct {
	torrentClient *torrent.Client
}

func (td *TorrentDownloader) Matches(url string) bool {
	return strings.HasPrefix(url, "magnet:")
}

func (td *TorrentDownloader) GetMetadata(ctx context.Context, url string) (*Metadata, error) {
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
	files := make([]FileMetadata, len(info.Files))
	for i, f := range info.Files {
		files[i] = FileMetadata{
			Path:     f.DisplayPath(info),
			LenBytes: f.Length,
		}
	}
	metadata := &Metadata{
		URL:   url,
		Name:  info.BestName(),
		Files: files,
	}
	return metadata, nil
}
