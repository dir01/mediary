package torrent_test

import (
	"bytes"
	"context"
	"fmt"
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

func TestDownload_UnknownFile(t *testing.T) {
	dwn, magnetURL := newDownloader(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := dwn.Download(ctx, magnetURL, []string{"01.mp3", "missing.mp3"})
	if err == nil || !strings.Contains(err.Error(), "torrent has 1 of 2 requested files") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDownload_StaticPeer(t *testing.T) {
	dwn, err := torrentdownloader.New(t.TempDir(), logger, false)
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}
	piece := bytes.Repeat([]byte{1}, tests.SeederPieceLength)
	seeder, magnetURL := tests.SetupLocalSeeder(t, "static-peer", map[string][]byte{"01.mp3": piece})
	// Only the static peer address tells the downloader where the seeder is.
	// A host name, to prove that the downloader resolves it.
	addrs, err := torrentdownloader.ParseStaticPeers(fmt.Sprintf(" localhost:%d ,", seeder.LocalPort()))
	if err != nil {
		t.Fatalf("failed to parse static peers: %v", err)
	}
	dwn.AddStaticPeers(addrs)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fps, err := dwn.Download(ctx, magnetURL, []string{"01.mp3"})
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if stat, err := os.Stat(fps["01.mp3"]); err != nil || stat.Size() != tests.SeederPieceLength {
		t.Errorf("file is not complete: %v", err)
	}
}

func TestParseStaticPeers(t *testing.T) {
	addrs, err := torrentdownloader.ParseStaticPeers("192.0.2.10:6881, [2001:db8::1]:6881, seeder.example:6881")
	want := []string{"192.0.2.10:6881", "[2001:db8::1]:6881", "seeder.example:6881"}
	if err != nil || strings.Join(addrs, ",") != strings.Join(want, ",") {
		t.Errorf("unexpected result: %v, %v", addrs, err)
	}
	if addrs, err := torrentdownloader.ParseStaticPeers(""); err != nil || len(addrs) != 0 {
		t.Errorf("expected no peers, got: %v, %v", addrs, err)
	}
	for _, bad := range []string{"seeder.example", ":6881", "seeder.example:0", "seeder.example:http", "seeder.example:70000"} {
		if _, err := torrentdownloader.ParseStaticPeers(bad); err == nil {
			t.Errorf("expected an error for %q", bad)
		}
	}
}

// TestResolveStaticPeers verifies that an IP address is used as it is, with the zone
// of a link-local IPv6 address, and that a host name is resolved.
func TestResolveStaticPeers(t *testing.T) {
	dwn, err := torrentdownloader.New(t.TempDir(), logger, false)
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}
	addrs, err := torrentdownloader.ParseStaticPeers("[fe80::1%eth0]:6881, 192.0.2.10:6881, localhost:6882")
	if err != nil {
		t.Fatalf("failed to parse static peers: %v", err)
	}
	dwn.AddStaticPeers(addrs)

	got := dwn.ResolveStaticPeers(context.Background())
	if len(got) < 3 || got[0] != "[fe80::1%eth0]:6881" || got[1] != "192.0.2.10:6881" {
		t.Fatalf("unexpected peers: %v", got)
	}
	for _, addr := range got[2:] {
		if addr != "127.0.0.1:6882" && addr != "[::1]:6882" {
			t.Errorf("unexpected address for localhost: %v", addr)
		}
	}
}
