package mediary_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/dir01/mediary"
)

func TestTorrentDownloader_Matches(t *testing.T) {
	d, err := mediary.NewTorrentDownloader()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	t.Run("matches magnet url", func(t *testing.T) {
		if !d.Matches("magnet:...") {
			t.Fatalf("expected to match magnet url")
		}
	})

	t.Run("does not match other url", func(t *testing.T) {
		if d.Matches("https://...") {
			t.Fatalf("expected to not match non-magnet url")
		}
	})
}

func TestTorrentDownloader_GetMetadata(t *testing.T) {
	d, err := mediary.NewTorrentDownloader()

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	t.Run("returns metadata", func(t *testing.T) {
		magnetUrl := "magnet:?xt=urn:btih:404E717C329D1B81F81FEFF82F119C6C53E1F3AE&dn=Mozart%20-%20The%20Very%20Best%20Of%20Mozart%20%5B2CDs%5D.&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A6969%2Fannounce&tr=udp%3A%2F%2F9.rarbg.to%3A2710%2Fannounce&tr=udp%3A%2F%2F9.rarbg.me%3A2780%2Fannounce&tr=udp%3A%2F%2F9.rarbg.to%3A2730%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=http%3A%2F%2Fp4p.arenabg.com%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.torrent.eu.org%3A451%2Fannounce&tr=udp%3A%2F%2Ftracker.tiny-vps.com%3A6969%2Fannounce&tr=udp%3A%2F%2Fopen.stealth.si%3A80%2Fannounce"
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		metadata, err := d.GetMetadata(ctx, magnetUrl)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				t.Skip("context deadline exceeded while getting metadata: either no internet or no seeders, skipping")
			}
			t.Fatalf("unexpected error: %s", err)
		}

		if metadata.URL != magnetUrl {
			t.Fatalf("expected url to be %s, got %s", magnetUrl, metadata.URL)
		}

		if metadata.Name != "Mozart - The Very Best Of Mozart [2CDs].www.lokotorrents.com" {
			t.Fatalf("expected name to be 'Mozart - The Very Best Of Mozart [2CDs]', got %s", metadata.Name)
		}

		expectedFiles := []mediary.FileMetadata{
			{Path: "CD1/01 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 14596223},
			{Path: "CD1/02 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 8627766},
			{Path: "CD1/03 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 7793937},
			{Path: "CD1/04 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 16097741},
			{Path: "CD1/05 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 10938035},
			{Path: "CD1/06 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 18000500},
			{Path: "CD1/07 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 10311096},
			{Path: "CD1/08 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 7631978},
			{Path: "CD1/09 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 8153382},
			{Path: "CD1/10 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 18446672},
			{Path: "CD1/11 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 6876517},
			{Path: "CD1/12 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 18478019},
			{Path: "CD1/13 из Mozart - The Very Best Of Mozart - CD1.mp3", LenBytes: 7023847},
			{Path: "CD2/01 из Mozart - The Very Best Of Mozart - CD2.mp3", LenBytes: 9835668},
			{Path: "CD2/02 из Mozart - The Very Best Of Mozart - CD2.mp3", LenBytes: 17748680},
			{Path: "CD2/03 из Mozart - The Very Best Of Mozart - CD2.mp3", LenBytes: 16524059},
			{Path: "CD2/04 из Mozart - The Very Best Of Mozart - CD2.mp3", LenBytes: 16360010},
			{Path: "CD2/05 из Mozart - The Very Best Of Mozart - CD2.mp3", LenBytes: 22840468},
			{Path: "CD2/06 из Mozart - The Very Best Of Mozart - CD2.mp3", LenBytes: 7087586},
			{Path: "CD2/07 из Mozart - The Very Best Of Mozart - CD2.mp3", LenBytes: 9029006},
			{Path: "CD2/08 из Mozart - The Very Best Of Mozart - CD2.mp3", LenBytes: 21548974},
			{Path: "CD2/09 из Mozart - The Very Best Of Mozart - CD2.mp3", LenBytes: 21286704},
			{Path: "CD2/10 из Mozart - The Very Best Of Mozart - CD2.mp3", LenBytes: 14549202},
			{Path: "CD2/11 из Mozart - The Very Best Of Mozart - CD2.mp3", LenBytes: 22429823},
		}
		if !reflect.DeepEqual(metadata.Files, expectedFiles) {
			t.Fatalf("expected files to be %v, got %v", expectedFiles, metadata.Files)
		}
	})
}
