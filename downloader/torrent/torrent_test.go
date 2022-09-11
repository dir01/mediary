package torrent_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/dir01/mediary/downloader/torrent"
	"github.com/dir01/mediary/service"
	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()

const magnetUrl = "magnet:?xt=urn:btih:ED0863F05DD89F56BC3E7A7242D72F044FD8B651"

func TestTorrentDownloader_AcceptsURL(t *testing.T) {
	d, err := torrent.NewTorrentDownloader("/nonexistent/temp/dir", logger)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	t.Run("matches magnet url", func(t *testing.T) {
		if !d.AcceptsURL(magnetUrl) {
			t.Fatalf("expected to match magnet url")
		}
	})

	t.Run("does not match other url", func(t *testing.T) {
		if d.AcceptsURL("https://...") {
			t.Fatalf("expected to not match non-magnet url")
		}
	})
}

func TestTorrentDownloader_GetMetadata(t *testing.T) {
	d, err := torrent.NewTorrentDownloader("/nonexistent/temp/dir", logger)

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	t.Run("returns metadata", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

		expectedName := "The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina & More"
		if metadata.Name != expectedName {
			t.Fatalf("expected name to be '%s', got %s", expectedName, metadata.Name)
		}

		var expectedFiles []service.FileMetadata
		if err = json.Unmarshal([]byte(`[
          {
            "path": "01 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 288670
          },
          {
            "path": "02 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 27625054
          },
          {
            "path": "03 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 27378718
          },
          {
            "path": "04 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 27484318
          },
          {
            "path": "05 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 27190750
          },
          {
            "path": "06 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 27052318
          },
          {
            "path": "07 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 27448222
          },
          {
            "path": "08 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 27563806
          },
          {
            "path": "09 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 27342814
          },
          {
            "path": "10 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 27526944
          },
          {
            "path": "11 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 26691552
          },
          {
            "path": "12 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 23577696
          },
          {
            "path": "13 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 24945120
          },
          {
            "path": "14 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 26514528
          },
          {
            "path": "15 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 26218464
          },
          {
            "path": "16 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 26722656
          },
          {
            "path": "17 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 26435808
          },
          {
            "path": "18 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 6373152
          },
          {
            "path": "19 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 6260448
          },
          {
            "path": "20 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 6155808
          },
          {
            "path": "21 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 5955936
          },
          {
            "path": "22 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 6207840
          },
          {
            "path": "23 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 6185376
          },
          {
            "path": "24 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 6309792
          },
          {
            "path": "25 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 6108192
          },
          {
            "path": "26 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 6223200
          },
          {
            "path": "27 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 6396576
          },
          {
            "path": "28 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 33316704
          },
          {
            "path": "29 - The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.mp3",
            "length_bytes": 13456800
          },
          {
            "path": "Leo Tolstoy BBC Radio Drama Collection.txt",
            "length_bytes": 2870
          },
          {
            "path": "The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.cue",
            "length_bytes": 1760
          },
          {
            "path": "The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.jpg",
            "length_bytes": 48462
          },
          {
            "path": "The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.nfo",
            "length_bytes": 1408
          }
		]`), &expectedFiles); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if !reflect.DeepEqual(metadata.Files, expectedFiles) {
			bytes, _ := json.MarshalIndent(metadata.Files, "", "  ")
			t.Fatalf("expected files to be %v, got %v", expectedFiles, string(bytes))
		}
	})
}

func TestTorrentDownloader_Download(t *testing.T) {
	tempDir := os.TempDir()
	defer func() { _ = os.RemoveAll(tempDir) }()

	d, err := torrent.NewTorrentDownloader(tempDir, logger)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filepathsMap, err := d.Download(ctx, magnetUrl, []string{
		"Leo Tolstoy BBC Radio Drama Collection.txt",
		"The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.cue",
		"The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina \u0026 More.nfo",
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	expectedPath := func(relPath string) (fullpath string) {
		return filepath.Join(
			tempDir,
			"The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina & More",
			relPath,
		)
	}
	expectedFilepathsMap := map[string]string{
		"Leo Tolstoy BBC Radio Drama Collection.txt": expectedPath("Leo Tolstoy BBC Radio Drama Collection.txt"),
		"The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina & More.cue": expectedPath("The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina & More.cue"),
		"The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina & More.nfo": expectedPath("The Leo Tolstoy BBC Radio Drama Collection Full-Cast Dramatisations of War and Peace, Anna Karenina & More.nfo"),
	}

	if !reflect.DeepEqual(filepathsMap, expectedFilepathsMap) {
		t.Fatalf("expected filepaths map to be %v, got %v", expectedFilepathsMap, filepathsMap)
	}
}
