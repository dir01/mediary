package tests

import (
	"os"
	"path/filepath"
	"testing"

	anacrolixTorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

// SetupLocalSeeder starts a local torrent seeder with the given files under torrentName/.
// Returns the seeder client (to pass to downloader.AddBootstrapPeer) and a canonical magnet URL.
// The magnet URL contains only the infohash — no trackers or x.pe hints — so it is stable
// across runs. Wire the seeder into the downloader via AddBootstrapPeer for direct peer exchange.
func SetupLocalSeeder(t *testing.T, torrentName string, files map[string][]byte) (*anacrolixTorrent.Client, string) {
	t.Helper()

	dir := t.TempDir()
	contentDir := filepath.Join(dir, torrentName)
	if err := os.MkdirAll(contentDir, 0o755); err != nil {
		t.Fatalf("SetupLocalSeeder: mkdir: %v", err)
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(contentDir, name), data, 0o644); err != nil {
			t.Fatalf("SetupLocalSeeder: write %s: %v", name, err)
		}
	}

	info := metainfo.Info{PieceLength: 256 * 1024}
	if err := info.BuildFromFilePath(contentDir); err != nil {
		t.Fatalf("SetupLocalSeeder: build metainfo: %v", err)
	}
	infoBytes, err := bencode.Marshal(info)
	if err != nil {
		t.Fatalf("SetupLocalSeeder: marshal info: %v", err)
	}
	mi := metainfo.MetaInfo{InfoBytes: infoBytes}

	cfg := anacrolixTorrent.NewDefaultClientConfig()
	cfg.DataDir = dir
	cfg.Seed = true
	cfg.NoDHT = true
	cfg.DisableTrackers = true
	cfg.ListenPort = 0

	client, err := anacrolixTorrent.NewClient(cfg)
	if err != nil {
		t.Fatalf("SetupLocalSeeder: new client: %v", err)
	}
	t.Cleanup(func() { client.Close() })

	torr, err := client.AddTorrent(&mi)
	if err != nil {
		t.Fatalf("SetupLocalSeeder: add torrent: %v", err)
	}
	torr.VerifyData()

	hash := mi.HashInfoBytes()
	magnetURL := metainfo.Magnet{InfoHash: hash}.String()
	return client, magnetURL
}
