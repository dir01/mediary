package media_processor

import (
	"context"
	"io"
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

func copyTestMP3(t *testing.T) string {
	t.Helper()
	src, err := os.Open("/root/go/pkg/mod/github.com/bogem/id3v2/v2@v2.1.4/testdata/test.mp3")
	if err != nil {
		t.Fatalf("failed to open test mp3: %v", err)
	}
	defer src.Close()

	tmp, err := os.CreateTemp("", "chapter_test_*.mp3")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := io.Copy(tmp, src); err != nil {
		os.Remove(tmp.Name())
		t.Fatalf("failed to copy test mp3: %v", err)
	}
	tmp.Close()
	t.Cleanup(func() { os.Remove(tmp.Name()) })
	return tmp.Name()
}

// copyTestMP3WithoutTag creates a copy of the test MP3 file with the ID3v2 tag
// stripped, simulating the output of ffmpeg concat (which typically has no tag).
func copyTestMP3WithoutTag(t *testing.T) string {
	t.Helper()
	src, err := os.Open("/root/go/pkg/mod/github.com/bogem/id3v2/v2@v2.1.4/testdata/test.mp3")
	if err != nil {
		t.Fatalf("failed to open test mp3: %v", err)
	}
	defer src.Close()

	// Read the ID3v2 header to find the tag size
	header := make([]byte, 10)
	if _, err := io.ReadFull(src, header); err != nil {
		t.Fatalf("failed to read header: %v", err)
	}
	if string(header[:3]) != "ID3" {
		t.Fatal("test file does not have ID3 header")
	}
	// Synchsafe size from bytes 6-9
	var tagSize int64
	for _, b := range header[6:10] {
		tagSize = (tagSize << 7) | int64(b&0x7F)
	}
	// Skip past the tag (header + body)
	if _, err := src.Seek(10+tagSize, io.SeekStart); err != nil {
		t.Fatalf("failed to seek past tag: %v", err)
	}

	tmp, err := os.CreateTemp("", "chapter_notag_*.mp3")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := io.Copy(tmp, src); err != nil {
		os.Remove(tmp.Name())
		t.Fatalf("failed to copy audio data: %v", err)
	}
	tmp.Close()
	t.Cleanup(func() { os.Remove(tmp.Name()) })
	return tmp.Name()
}

func TestAddChapterTags_MultipleChapters(t *testing.T) {
	filepath := copyTestMP3(t)

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
	defer tag.Close()

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
	filepath := copyTestMP3WithoutTag(t)

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
	defer tag.Close()

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
