package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	anacrolixTorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

// MakeMinimalMP3 generates a short valid MP3 file using FFmpeg and returns its bytes.
// The resulting file is a 0.1-second mono sine wave — small but processable by FFmpeg.
func MakeMinimalMP3(t *testing.T) []byte {
	t.Helper()
	outFile := filepath.Join(t.TempDir(), "audio.mp3")
	cmd := exec.Command("ffmpeg", "-y",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=0.1",
		"-ar", "44100", "-ac", "1", "-b:a", "32k",
		outFile,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg failed to generate test MP3: %v\n%s", err, out)
	}
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("MakeMinimalMP3: read: %v", err)
	}
	return data
}

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
	_ = torr.VerifyDataContext(t.Context())

	hash := mi.HashInfoBytes()
	magnetURL := metainfo.Magnet{InfoHash: hash}.String()
	return client, magnetURL
}
