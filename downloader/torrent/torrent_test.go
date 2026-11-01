package torrent_test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	torrentdownloader "github.com/dir01/mediary/downloader/torrent"
	"github.com/dir01/mediary/tests"
)

var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

// newDownloader returns a downloader that peers with a seeder that has
// "01.mp3" and "02.mp3", but not "03.mp3".
func newDownloader(t *testing.T) (*torrentdownloader.Downloader, string) {
	t.Helper()
	dwn, err := torrentdownloader.New(t.TempDir(), logger, false)
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}
	piece := bytes.Repeat([]byte{1}, tests.SeederPieceLength)
	seeder, magnetURL := tests.SetupPartialSeeder(t, "partial-collection", map[string][]byte{
		"01.mp3": piece,
		"02.mp3": piece,
		"03.mp3": piece,
	}, []string{"03.mp3"})
	dwn.AddBootstrapPeer(seeder)
	return dwn, magnetURL
}

func TestDownload_CompleteFiles(t *testing.T) {
	dwn, magnetURL := newDownloader(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fps, err := dwn.Download(ctx, magnetURL, []string{"01.mp3", "02.mp3"})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	for _, name := range []string{"01.mp3", "02.mp3"} {
		stat, err := os.Stat(fps[name])
		if err != nil || stat.Size() != tests.SeederPieceLength {
			t.Errorf("file %s is not complete: %v", name, err)
		}
	}
}

func TestDownload_TimeoutWithIncompleteFiles(t *testing.T) {
	dwn, magnetURL := newDownloader(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fps, err := dwn.Download(ctx, magnetURL, []string{"01.mp3", "02.mp3", "03.mp3"})
	if err == nil {
		t.Fatalf("expected an error, got success: %v", fps)
	}
	if !strings.Contains(err.Error(), "download timed out: 2 of 3 files complete (512 KB of 768 KB)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDownload_Stall(t *testing.T) {
	dwn, magnetURL := newDownloader(t)
	dwn.SetStallTimeout(3 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := dwn.Download(ctx, magnetURL, []string{"01.mp3", "03.mp3"})
	if err == nil {
		t.Fatal("expected an error, got success")
	}
	if !strings.Contains(err.Error(), "download stalled: no data received in 3s: 1 of 2 files complete") {
		t.Errorf("unexpected error: %v", err)
	}
	if ctx.Err() != nil {
		t.Errorf("expected the stall to stop the download before the context deadline")
	}
}
