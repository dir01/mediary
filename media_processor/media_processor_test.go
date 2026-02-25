package media_processor

import (
	"context"
	"os"
	"testing"
	"time"

	id3v2 "github.com/bogem/id3v2/v2"
	"github.com/dir01/mediary/service"
	"go.uber.org/zap"
)

func TestConcatenate_EmptyFilepathsReturnsError(t *testing.T) {
	processor := &FFMpegMediaProcessor{log: zap.NewNop()}

	_, err := processor.Concatenate(context.Background(), nil, "aac")
	if err == nil {
		t.Fatal("expected error for empty filepaths")
	}
}

func TestConcatenate_FirstFilepathWithoutExtensionReturnsError(t *testing.T) {
	processor := &FFMpegMediaProcessor{log: zap.NewNop()}

	_, err := processor.Concatenate(context.Background(), []string{"/tmp/audio"}, "aac")
	if err == nil {
		t.Fatal("expected error when first filepath has no extension")
	}
}

// createTestMP3WithTag creates a temp file containing a minimal ID3v2 tag
// followed by dummy audio bytes. This simulates a real MP3 with an existing tag.
func createTestMP3WithTag(t *testing.T) string {
	t.Helper()

	// Start with dummy audio bytes (the id3v2 library treats everything
	// after the tag as opaque data, so actual MP3 validity doesn't matter).
	tmp, err := os.CreateTemp("", "chapter_test_*.mp3")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	// Write some dummy audio bytes
	if _, err := tmp.Write(make([]byte, 512)); err != nil {
		_ = os.Remove(tmp.Name())
		t.Fatalf("failed to write dummy audio: %v", err)
	}
	_ = tmp.Close()
	t.Cleanup(func() { _ = os.Remove(tmp.Name()) })

	// Write a minimal ID3v2 tag so the file has one before our test runs.
	tag, err := id3v2.Open(tmp.Name(), id3v2.Options{Parse: false})
	if err != nil {
		t.Fatalf("failed to open file for tagging: %v", err)
	}
	tag.SetVersion(4)
	tag.SetArtist("test")
	if err := tag.Save(); err != nil {
		t.Fatalf("failed to save seed tag: %v", err)
	}
	if err := tag.Close(); err != nil {
		t.Fatalf("failed to close tag: %v", err)
	}

	return tmp.Name()
}

// createTestMP3WithoutTag creates a temp file containing only dummy audio
// bytes (no ID3 tag), simulating the typical output of ffmpeg concat.
func createTestMP3WithoutTag(t *testing.T) string {
	t.Helper()

	tmp, err := os.CreateTemp("", "chapter_notag_*.mp3")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := tmp.Write(make([]byte, 512)); err != nil {
		_ = os.Remove(tmp.Name())
		t.Fatalf("failed to write dummy audio: %v", err)
	}
	_ = tmp.Close()
	t.Cleanup(func() { _ = os.Remove(tmp.Name()) })
	return tmp.Name()
}

func TestAddChapterTags_MultipleChapters(t *testing.T) {
	filepath := createTestMP3WithTag(t)

	chapters := []service.Chapter{
		{Title: "Intro", StartTime: 0, EndTime: 5 * time.Minute},
		{Title: "Chapter 1", StartTime: 5 * time.Minute, EndTime: 15 * time.Minute},
		{Title: "Chapter 2", StartTime: 15 * time.Minute, EndTime: 30 * time.Minute},
	}

	processor := &FFMpegMediaProcessor{log: zap.NewNop()}
	if err := processor.AddChapterTags(context.Background(), filepath, chapters); err != nil {
		t.Fatalf("AddChapterTags failed: %v", err)
	}

	// Read back and verify
	tag, err := id3v2.Open(filepath, id3v2.Options{Parse: true})
	if err != nil {
		t.Fatalf("failed to open tagged file: %v", err)
	}
	defer func() { _ = tag.Close() }()

	frames := tag.GetFrames("CHAP")
	if len(frames) != len(chapters) {
		t.Fatalf("expected %d chapter frames, got %d", len(chapters), len(frames))
	}

	for i, f := range frames {
		cf, ok := f.(id3v2.ChapterFrame)
		if !ok {
			t.Fatalf("frame %d is not a ChapterFrame", i)
		}
		if cf.StartTime != chapters[i].StartTime {
			t.Errorf("chapter %d StartTime: want %v, got %v", i, chapters[i].StartTime, cf.StartTime)
		}
		if cf.EndTime != chapters[i].EndTime {
			t.Errorf("chapter %d EndTime: want %v, got %v", i, chapters[i].EndTime, cf.EndTime)
		}
		if cf.Title == nil || cf.Title.Text != chapters[i].Title {
			got := ""
			if cf.Title != nil {
				got = cf.Title.Text
			}
			t.Errorf("chapter %d Title: want %q, got %q", i, chapters[i].Title, got)
		}
	}
}

// TestAddChapterTags_FileWithoutExistingTag verifies that chapters are
// written correctly to a file that has no pre-existing ID3 tag, which is
// the typical case for ffmpeg-concatenated output.
func TestAddChapterTags_FileWithoutExistingTag(t *testing.T) {
	filepath := createTestMP3WithoutTag(t)

	chapters := []service.Chapter{
		{Title: "Intro", StartTime: 0, EndTime: 1 * time.Minute},
		{Title: "Chapter 1", StartTime: 1 * time.Minute, EndTime: 3 * time.Minute},
		{Title: "Chapter 2", StartTime: 3 * time.Minute, EndTime: 5 * time.Minute},
	}

	processor := &FFMpegMediaProcessor{log: zap.NewNop()}
	if err := processor.AddChapterTags(context.Background(), filepath, chapters); err != nil {
		t.Fatalf("AddChapterTags failed: %v", err)
	}

	// Read back and verify
	tag, err := id3v2.Open(filepath, id3v2.Options{Parse: true})
	if err != nil {
		t.Fatalf("failed to open tagged file: %v", err)
	}
	defer func() { _ = tag.Close() }()

	frames := tag.GetFrames("CHAP")
	if len(frames) != len(chapters) {
		t.Fatalf("expected %d chapter frames, got %d", len(chapters), len(frames))
	}

	for i, f := range frames {
		cf, ok := f.(id3v2.ChapterFrame)
		if !ok {
			t.Fatalf("frame %d is not a ChapterFrame", i)
		}
		if cf.StartTime != chapters[i].StartTime {
			t.Errorf("chapter %d StartTime: want %v, got %v", i, chapters[i].StartTime, cf.StartTime)
		}
		if cf.EndTime != chapters[i].EndTime {
			t.Errorf("chapter %d EndTime: want %v, got %v", i, chapters[i].EndTime, cf.EndTime)
		}
		if cf.Title == nil || cf.Title.Text != chapters[i].Title {
			got := ""
			if cf.Title != nil {
				got = cf.Title.Text
			}
			t.Errorf("chapter %d Title: want %q, got %q", i, chapters[i].Title, got)
		}
	}
}
