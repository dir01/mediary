package service_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/service/mocks"
	"github.com/gojuno/minimock/v3"
)

// TestConcatenateFlow_ZeroDurationSkipsChapters verifies that when GetInfo
// returns Duration=0, chapter tags are skipped instead of being written with
// all-zero timestamps. This was caused by a bug in GetDuration where
// zaperr.Wrap(nil, ...) silently returned nil instead of an error, leading
// to Duration=0 propagating without any error signal.
func TestConcatenateFlow_ZeroDurationSkipsChapters(t *testing.T) {
	mc := minimock.NewController(t)

	storage := mocks.NewStorageMock(mc)
	queue := mocks.NewJobsQueueMock(mc)
	dwn := mocks.NewDownloaderMock(mc)
	mp := mocks.NewMediaProcessorMock(mc)
	upl := mocks.NewUploaderMock(mc)

	var onJob func(payloadBytes []byte) error
	queue.SubscribeMock.Set(func(_ context.Context, _ string, f func([]byte) error) {
		onJob = f
	})
	queue.RunMock.Set(func() {})
	queue.ShutdownMock.Set(func() {})

	svc := service.NewService(dwn, storage, queue, mp, upl, logger)
	svc.Start()
	defer svc.Stop()

	jobID := "test-job-zero-dur"
	job := &service.Job{
		JobParams: service.JobParams{
			URL:  "http://example.com/audio",
			Type: "concatenate",
			Params: map[string]interface{}{
				"variants":   []interface{}{"intro.mp3", "chapter1.mp3", "chapter2.mp3"},
				"audioCodec": "copy",
				"uploadUrl":  "http://example.com/upload",
			},
		},
		ID:            jobID,
		DisplayStatus: "created",
	}
	storage.GetJobMock.Set(func(_ context.Context, id string) (*service.Job, error) {
		return job, nil
	})
	storage.SaveJobMock.Set(func(_ context.Context, _ *service.Job) error {
		return nil
	})

	fpMap := map[string]string{
		"intro.mp3":    "/tmp/dl/intro.mp3",
		"chapter1.mp3": "/tmp/dl/chapter1.mp3",
		"chapter2.mp3": "/tmp/dl/chapter2.mp3",
	}
	dwn.AcceptsURLMock.Optional().Return(true)
	dwn.DownloadMock.Set(func(_ context.Context, url string, fps []string) (map[string]string, error) {
		return fpMap, nil
	})

	resultPath := "/tmp/result/output.mp3"

	// GetInfo returns Duration=0 with no error. Before the fix, this
	// caused chapters with all-zero timestamps to be written.
	mp.GetInfoMock.Set(func(_ context.Context, fp string) (*service.MediaInfo, error) {
		return &service.MediaInfo{Duration: 0, FileLenBytes: 1024}, nil
	})

	mp.ConcatenateMock.Set(func(_ context.Context, fps []string, codec string) (string, error) {
		return resultPath, nil
	})

	// AddChapterTags should NOT be called when durations are zero.
	addChapterTagsCalled := false
	mp.AddChapterTagsMock.Optional().Set(func(_ context.Context, fp string, chapters []service.Chapter) error {
		addChapterTagsCalled = true
		return nil
	})

	upl.UploadMock.Set(func(_ context.Context, fp string, url string) error {
		return nil
	})

	payload, _ := json.Marshal(jobID)
	if err := onJob(payload); err != nil {
		t.Fatalf("onJob failed: %v", err)
	}

	if addChapterTagsCalled {
		t.Error("AddChapterTags should not be called when file durations are zero")
	}
}

func TestConcatenateFlow_ChapterTimestamps(t *testing.T) {
	mc := minimock.NewController(t)

	storage := mocks.NewStorageMock(mc)
	queue := mocks.NewJobsQueueMock(mc)
	dwn := mocks.NewDownloaderMock(mc)
	mp := mocks.NewMediaProcessorMock(mc)
	upl := mocks.NewUploaderMock(mc)

	// Capture the queue subscriber callback so we can invoke it directly.
	var onJob func(payloadBytes []byte) error
	queue.SubscribeMock.Set(func(_ context.Context, _ string, f func([]byte) error) {
		onJob = f
	})
	queue.RunMock.Set(func() {})
	queue.ShutdownMock.Set(func() {})

	svc := service.NewService(dwn, storage, queue, mp, upl, logger)
	svc.Start()
	defer svc.Stop()

	// --- set up the job -------------------------------------------------------
	jobID := "test-job-chapters"
	jobURL := "http://example.com/audio"

	job := &service.Job{
		JobParams: service.JobParams{
			URL:  jobURL,
			Type: "concatenate",
			Params: map[string]interface{}{
				"variants":   []interface{}{"intro.mp3", "chapter1.mp3", "chapter2.mp3"},
				"audioCodec": "copy",
				"uploadUrl":  "http://example.com/upload",
			},
		},
		ID:            jobID,
		DisplayStatus: "created",
	}

	// Storage.GetJob is called twice: once from onPublishedJob, once inside the flow.
	storage.GetJobMock.Set(func(_ context.Context, id string) (*service.Job, error) {
		if id != jobID {
			t.Fatalf("unexpected job id: %s", id)
		}
		return job, nil
	})
	storage.SaveJobMock.Set(func(_ context.Context, _ *service.Job) error {
		return nil
	})

	// Downloader returns a deterministic filepath mapping.
	variants := []string{"intro.mp3", "chapter1.mp3", "chapter2.mp3"}
	fpMap := map[string]string{
		"intro.mp3":    "/tmp/dl/intro.mp3",
		"chapter1.mp3": "/tmp/dl/chapter1.mp3",
		"chapter2.mp3": "/tmp/dl/chapter2.mp3",
	}
	dwn.AcceptsURLMock.Optional().Return(true)
	dwn.DownloadMock.Set(func(_ context.Context, url string, fps []string) (map[string]string, error) {
		return fpMap, nil
	})

	// GetInfo returns known durations for individual files and for the result.
	fileDurations := map[string]time.Duration{
		"/tmp/dl/intro.mp3":    1 * time.Minute,
		"/tmp/dl/chapter1.mp3": 2 * time.Minute,
		"/tmp/dl/chapter2.mp3": 90 * time.Second,
	}
	resultPath := "/tmp/result/output.mp3"
	mp.GetInfoMock.Set(func(_ context.Context, fp string) (*service.MediaInfo, error) {
		if d, ok := fileDurations[fp]; ok {
			return &service.MediaInfo{Duration: d, FileLenBytes: 1024}, nil
		}
		// result file
		return &service.MediaInfo{Duration: 270 * time.Second, FileLenBytes: 3072}, nil
	})

	mp.ConcatenateMock.Set(func(_ context.Context, fps []string, codec string) (string, error) {
		return resultPath, nil
	})

	// Capture chapters passed to AddChapterTags.
	var gotChapters []service.Chapter
	mp.AddChapterTagsMock.Set(func(_ context.Context, fp string, chapters []service.Chapter) error {
		if fp != resultPath {
			t.Errorf("AddChapterTags called on unexpected file: %s", fp)
		}
		gotChapters = chapters
		return nil
	})

	upl.UploadMock.Set(func(_ context.Context, fp string, url string) error {
		return nil
	})

	// --- execute the flow via the captured callback ---
	payload, _ := json.Marshal(jobID)
	if err := onJob(payload); err != nil {
		t.Fatalf("onJob failed: %v", err)
	}

	// --- assertions -----------------------------------------------------------
	if len(gotChapters) != len(variants) {
		t.Fatalf("expected %d chapters, got %d", len(variants), len(gotChapters))
	}

	wantChapters := []service.Chapter{
		{Title: "intro", StartTime: 0, EndTime: 1 * time.Minute},
		{Title: "chapter1", StartTime: 1 * time.Minute, EndTime: 3 * time.Minute},
		{Title: "chapter2", StartTime: 3 * time.Minute, EndTime: 4*time.Minute + 30*time.Second},
	}

	for i, want := range wantChapters {
		got := gotChapters[i]
		if got.Title != want.Title {
			t.Errorf("chapter %d Title: want %q, got %q", i, want.Title, got.Title)
		}
		if got.StartTime != want.StartTime {
			t.Errorf("chapter %d StartTime: want %v, got %v", i, want.StartTime, got.StartTime)
		}
		if got.EndTime != want.EndTime {
			t.Errorf("chapter %d EndTime: want %v, got %v", i, want.EndTime, got.EndTime)
		}
	}
}
