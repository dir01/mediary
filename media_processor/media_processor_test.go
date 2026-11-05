package media_processor

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	id3v2 "github.com/bogem/id3v2/v2"
	"github.com/dir01/mediary/service"
)

var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestParseDurationOutput(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    float64
		wantErr bool
	}{
		{
			name:   "clean output",
			output: "22930.241438\n",
			want:   22930.241438,
		},
		{
			name:   "bmp bad magic number warning before value",
			output: "[bmp @ 0xffffa5eaeb40] bad magic number\n22930.241438\n",
			want:   22930.241438,
		},
		{
			name:   "multiple warning lines before value",
			output: "[foo @ 0x1] some warning\n[bar @ 0x2] another warning\n123.456\n",
			want:   123.456,
		},
		{
			name:    "no valid float",
			output:  "[bmp @ 0x1] bad magic number\n",
			wantErr: true,
		},
		{
			name:    "empty output",
			output:  "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDurationOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseDurationOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("parseDurationOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConcatenate_EmptyFilepathsReturnsError(t *testing.T) {
	processor := &FFMpegMediaProcessor{log: testLogger}

	_, err := processor.Concatenate(context.Background(), nil, "aac")
	if err == nil {
		t.Fatal("expected error for empty filepaths")
	}
}

func TestConcatenate_FirstFilepathWithoutExtensionReturnsError(t *testing.T) {
	processor := &FFMpegMediaProcessor{log: testLogger}

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

	processor := &FFMpegMediaProcessor{log: testLogger}
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

	processor := &FFMpegMediaProcessor{log: testLogger}
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

// makeTestMP3 generates a short valid MP3 file with the given bitrate using FFmpeg.
func makeTestMP3(t *testing.T, duration float64, bitrate string) string {
	t.Helper()
	outFile := filepath.Join(t.TempDir(), "audio.mp3")
	cmd := exec.Command("ffmpeg", "-y",
		"-f", "lavfi", "-i", fmt.Sprintf("sine=frequency=440:duration=%g", duration),
		"-ar", "44100", "-ac", "1", "-b:a", bitrate,
		outFile,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg failed to generate test MP3: %v\n%s", err, out)
	}
	return outFile
}

// TestConcatenate_OutputSizeNotLargerThanInputs verifies that concatenating
// MP3 files with re-encoding does not produce output larger than the sum of inputs.
// This reproduces the bug where ffmpeg used its default 128kbps bitrate instead of
// matching the source bitrate, inflating the output.
func TestConcatenate_OutputSizeNotLargerThanInputs(t *testing.T) {
	// Create two MP3 files at 32kbps — a typical audiobook bitrate.
	file1 := makeTestMP3(t, 1.0, "32k")
	file2 := makeTestMP3(t, 1.0, "32k")

	stat1, err := os.Stat(file1)
	if err != nil {
		t.Fatalf("stat file1: %v", err)
	}
	stat2, err := os.Stat(file2)
	if err != nil {
		t.Fatalf("stat file2: %v", err)
	}
	inputTotal := stat1.Size() + stat2.Size()

	processor := &FFMpegMediaProcessor{log: testLogger}
	resultPath, err := processor.Concatenate(context.Background(), []string{file1, file2}, "mp3")
	if err != nil {
		t.Fatalf("Concatenate failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(resultPath) })

	resultStat, err := os.Stat(resultPath)
	if err != nil {
		t.Fatalf("stat result: %v", err)
	}

	// Allow 10% overhead for container/header differences, but no more.
	maxExpected := int64(float64(inputTotal) * 1.10)
	if resultStat.Size() > maxExpected {
		t.Errorf("concatenated file is too large: result=%d bytes, inputs total=%d bytes (max allowed=%d). "+
			"This suggests ffmpeg is re-encoding at a higher bitrate than the source files.",
			resultStat.Size(), inputTotal, maxExpected)
	}
}

// makeTestMP3WithCover returns an MP3 file with a JPEG cover picture, like the
// files of many audiobooks.
func makeTestMP3WithCover(t *testing.T) string {
	t.Helper()
	audioFile := makeTestMP3(t, 5.0, "32k")
	coverFile := filepath.Join(t.TempDir(), "cover.jpg")
	outFile := filepath.Join(t.TempDir(), "with_cover.mp3")
	for _, args := range [][]string{
		{"-y", "-f", "lavfi", "-i", "color=c=red:s=64x64", "-frames:v", "1", coverFile},
		{"-y", "-i", audioFile, "-i", coverFile, "-map", "0:a", "-map", "1:v", "-c", "copy",
			"-disposition:v", "attached_pic",
			"-metadata", "title=Test Title", "-metadata", "artist=Test Artist", outFile},
	} {
		if out, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
			t.Fatalf("ffmpeg failed to generate test MP3 with cover: %v\n%s", err, out)
		}
	}
	return outFile
}

// TestConcatenate_CopiesCoverPicture verifies that Concatenate copies the cover
// picture and does not encode it again. When ffmpeg encodes the cover, the mp3
// muxer keeps all the audio in memory until the end of the input.
func TestConcatenate_CopiesCoverPicture(t *testing.T) {
	file := makeTestMP3WithCover(t)
	processor := &FFMpegMediaProcessor{log: testLogger}

	for _, audioCodec := range []string{"copy", "mp3"} {
		t.Run(audioCodec, func(t *testing.T) {
			resultPath, err := processor.Concatenate(context.Background(), []string{file, file}, audioCodec)
			if err != nil {
				t.Fatalf("Concatenate failed: %v", err)
			}
			t.Cleanup(func() { _ = os.Remove(resultPath) })

			out, err := exec.Command("ffprobe", "-v", "error",
				"-show_entries", "stream=codec_type,codec_name",
				"-of", "csv=p=0", resultPath).Output()
			if err != nil {
				t.Fatalf("ffprobe failed: %v", err)
			}
			if got, want := string(out), "mp3,audio\nmjpeg,video\n"; got != want {
				t.Errorf("unexpected streams: got %q, want %q", got, want)
			}
		})
	}
}

// TestConcatenate_WithoutCoverPicture verifies that a file without a cover
// picture is not an error ("-map 0:v?" is optional).
func TestConcatenate_WithoutCoverPicture(t *testing.T) {
	file := makeTestMP3(t, 1.0, "32k")
	processor := &FFMpegMediaProcessor{log: testLogger}
	resultPath, err := processor.Concatenate(context.Background(), []string{file, file}, "copy")
	if err != nil {
		t.Fatalf("Concatenate failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(resultPath) })
}

// TestAddChapterTags_KeepsCoverAndTags verifies the order of the concatenate flow:
// Concatenate copies the cover picture, and AddChapterTags must keep it, together
// with the other tags of the input.
func TestAddChapterTags_KeepsCoverAndTags(t *testing.T) {
	file := makeTestMP3WithCover(t)
	processor := &FFMpegMediaProcessor{log: testLogger}
	resultPath, err := processor.Concatenate(context.Background(), []string{file, file}, "copy")
	if err != nil {
		t.Fatalf("Concatenate failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(resultPath) })

	chapters := []service.Chapter{
		{Title: "01", StartTime: 0, EndTime: 5 * time.Second},
		{Title: "02", StartTime: 5 * time.Second, EndTime: 10 * time.Second},
	}
	if err := processor.AddChapterTags(context.Background(), resultPath, chapters); err != nil {
		t.Fatalf("AddChapterTags failed: %v", err)
	}

	tag, err := id3v2.Open(resultPath, id3v2.Options{Parse: true})
	if err != nil {
		t.Fatalf("failed to open tagged file: %v", err)
	}
	defer func() { _ = tag.Close() }()
	if n := len(tag.GetFrames("APIC")); n != 1 {
		t.Errorf("expected 1 cover picture (APIC), got %d", n)
	}
	if tag.Title() != "Test Title" || tag.Artist() != "Test Artist" {
		t.Errorf("tags were lost: title %q, artist %q", tag.Title(), tag.Artist())
	}
	if n := len(tag.GetFrames("CHAP")); n != len(chapters) {
		t.Errorf("expected %d chapter frames, got %d", len(chapters), n)
	}
}

// TestAddChapterTags_ReplacesOldChapters verifies that chapters that the file
// already has are replaced, not added to.
func TestAddChapterTags_ReplacesOldChapters(t *testing.T) {
	filepath := createTestMP3WithTag(t)
	processor := &FFMpegMediaProcessor{log: testLogger}
	for _, chapters := range [][]service.Chapter{
		{{Title: "old 1", StartTime: 0, EndTime: time.Minute}, {Title: "old 2", StartTime: time.Minute, EndTime: 2 * time.Minute}},
		{{Title: "new", StartTime: 0, EndTime: 3 * time.Minute}},
	} {
		if err := processor.AddChapterTags(context.Background(), filepath, chapters); err != nil {
			t.Fatalf("AddChapterTags failed: %v", err)
		}
	}

	tag, err := id3v2.Open(filepath, id3v2.Options{Parse: true})
	if err != nil {
		t.Fatalf("failed to open tagged file: %v", err)
	}
	defer func() { _ = tag.Close() }()
	frames := tag.GetFrames("CHAP")
	if len(frames) != 1 {
		t.Fatalf("expected 1 chapter frame, got %d", len(frames))
	}
	if cf, ok := frames[0].(id3v2.ChapterFrame); !ok || cf.Title == nil || cf.Title.Text != "new" {
		t.Errorf("unexpected chapter frame: %+v", frames[0])
	}
	if n := len(tag.GetFrames("CTOC")); n != 1 {
		t.Errorf("expected 1 CTOC frame, got %d", n)
	}
	if tag.Artist() != "test" {
		t.Errorf("artist of the seed tag was lost: %q", tag.Artist())
	}
}

// makeTestMP4Audio returns an AAC file in an MP4 container (m4a or m4b), with a
// cover picture and a title tag. The file is in a directory with a single quote
// in its name, like "03 Abaddon's Gate".
func makeTestMP4Audio(t *testing.T, ext string, duration float64) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "Abaddon's Gate")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	coverFile := filepath.Join(dir, "cover.jpg")
	outFile := filepath.Join(dir, "part"+ext)
	for _, args := range [][]string{
		{"-y", "-f", "lavfi", "-i", "color=c=red:s=64x64", "-frames:v", "1", coverFile},
		{"-y", "-f", "lavfi", "-i", fmt.Sprintf("sine=frequency=440:duration=%g", duration),
			"-i", coverFile, "-map", "0:a", "-map", "1:v",
			"-c:a", "aac", "-b:a", "32k", "-c:v", "copy", "-disposition:v", "attached_pic",
			"-metadata", "title=Test Title", "-f", "mp4", outFile},
	} {
		if out, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
			t.Fatalf("ffmpeg failed to generate test %s: %v\n%s", ext, err, out)
		}
	}
	return outFile
}

// TestConcatenate_MP4AudioToMP3 verifies the concatenate flow for m4a and m4b
// files: the result is an MP3 file with all the audio, the cover picture, the
// tags and the chapters. Before, Concatenate read only the first MP4 file, and
// AddChapterTags made the result unreadable with an ID3 tag at its start.
func TestConcatenate_MP4AudioToMP3(t *testing.T) {
	for _, ext := range []string{".m4a", ".m4b"} {
		t.Run(ext, func(t *testing.T) {
			file1 := makeTestMP4Audio(t, ext, 2.0)
			file2 := makeTestMP4Audio(t, ext, 3.0)
			processor := &FFMpegMediaProcessor{log: testLogger}

			resultPath, err := processor.Concatenate(context.Background(), []string{file1, file2}, "copy")
			if err != nil {
				t.Fatalf("Concatenate failed: %v", err)
			}
			t.Cleanup(func() { _ = os.Remove(resultPath) })
			if filepath.Ext(resultPath) != ".mp3" {
				t.Errorf("expected an .mp3 result, got %q", resultPath)
			}

			chapters := []service.Chapter{
				{Title: "Part 1", StartTime: 0, EndTime: 2 * time.Second},
				{Title: "Part 2", StartTime: 2 * time.Second, EndTime: 5 * time.Second},
			}
			if err := processor.AddChapterTags(context.Background(), resultPath, chapters); err != nil {
				t.Fatalf("AddChapterTags failed: %v", err)
			}

			duration, err := processor.GetDuration(resultPath)
			if err != nil {
				t.Fatalf("GetDuration failed: %v", err)
			}
			if duration < 4900*time.Millisecond || duration > 5200*time.Millisecond {
				t.Errorf("expected duration of about 5s, got %v", duration)
			}

			out, err := exec.Command("ffprobe", "-v", "error",
				"-show_entries", "stream=codec_name", "-of", "csv=p=0", resultPath).Output()
			if err != nil {
				t.Fatalf("ffprobe failed: %v", err)
			}
			if got, want := string(out), "mp3\nmjpeg\n"; got != want {
				t.Errorf("unexpected streams: got %q, want %q", got, want)
			}

			tag, err := id3v2.Open(resultPath, id3v2.Options{Parse: true})
			if err != nil {
				t.Fatalf("failed to open tagged file: %v", err)
			}
			defer func() { _ = tag.Close() }()
			if tag.Title() != "Test Title" {
				t.Errorf("title tag was lost: got %q", tag.Title())
			}
			if n := len(tag.GetFrames("APIC")); n != 1 {
				t.Errorf("expected 1 cover picture (APIC), got %d", n)
			}
			if n := len(tag.GetFrames("CHAP")); n != len(chapters) {
				t.Errorf("expected %d chapter frames, got %d", len(chapters), n)
			}
		})
	}
}

// TestAddChapterTags_NotMP3ReturnsError verifies that AddChapterTags does not
// write an ID3 tag into a file that is not MP3, because the tag makes an MP4
// file unreadable.
func TestAddChapterTags_NotMP3ReturnsError(t *testing.T) {
	file := makeTestMP4Audio(t, ".m4a", 1.0)
	processor := &FFMpegMediaProcessor{log: testLogger}

	chapters := []service.Chapter{{Title: "01", StartTime: 0, EndTime: time.Second}}
	if err := processor.AddChapterTags(context.Background(), file, chapters); err == nil {
		t.Error("expected error for an m4a file")
	}
	if _, err := processor.GetDuration(file); err != nil {
		t.Errorf("file is not readable after AddChapterTags: %v", err)
	}
}

// TestConcatenate_MP4AudioWithChapters verifies that the MP3 result does not
// get the chapters of the first m4b file (they are only for that file), and
// that a 32 kbit/s AAC input gives an MP3 of at least minTranscodeBitrate.
func TestConcatenate_MP4AudioWithChapters(t *testing.T) {
	dir := t.TempDir()
	metadataFile := filepath.Join(dir, "chapters.txt")
	if err := os.WriteFile(metadataFile,
		[]byte(";FFMETADATA1\n[CHAPTER]\nTIMEBASE=1/1000\nSTART=0\nEND=1000\ntitle=Inner\n"), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	file := filepath.Join(dir, "book.m4b")
	if out, err := exec.Command("ffmpeg", "-y",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=2", "-i", metadataFile,
		"-map", "0:a", "-map_chapters", "1", "-c:a", "aac", "-b:a", "32k", "-f", "mp4", file,
	).CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg failed to generate test m4b: %v\n%s", err, out)
	}

	processor := &FFMpegMediaProcessor{log: testLogger}
	resultPath, err := processor.Concatenate(context.Background(), []string{file, file}, "copy")
	if err != nil {
		t.Fatalf("Concatenate failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(resultPath) })

	out, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "chapter=start_time",
		"-of", "csv=p=0", resultPath).Output()
	if err != nil {
		t.Fatalf("ffprobe failed: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected no chapters, got %q", out)
	}

	bitrate, err := processor.getAudioBitrate(resultPath)
	if err != nil {
		t.Fatalf("getAudioBitrate failed: %v", err)
	}
	if n, _ := strconv.Atoi(bitrate); n < minTranscodeBitrate {
		t.Errorf("expected a bitrate of at least %d, got %s", minTranscodeBitrate, bitrate)
	}
}
