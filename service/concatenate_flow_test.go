package service_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/service/mocks"
	"github.com/gojuno/minimock/v3"
)

// TestConcatenateFlow_ConcatDurationsErrorSkipsChapters verifies that when
// the durations for the chapters cannot be measured, chapter tags are skipped
// but concatenation still succeeds.
func TestConcatenateFlow_ConcatDurationsErrorSkipsChapters(t *testing.T) {
	mc := minimock.NewController(t)

	storage := mocks.NewStorageMock(mc)
	queue := mocks.NewJobsQueueMock(mc)
	dwn := mocks.NewDownloaderMock(mc)
	mp := mocks.NewMediaProcessorMock(mc)
	upl := mocks.NewUploaderMock(mc)

	var onJob func(ctx context.Context, payloadBytes []byte) error
	queue.SubscribeMock.Set(func(_ context.Context, _ string, f func(context.Context, []byte) error, _ func(context.Context, *sql.Tx, []byte, service.Attempt, error) error) {
		onJob = f
	})
	queue.RunMock.Set(func() {})
	queue.ShutdownMock.Set(func() {})

	svc := service.NewService(dwn, storage, queue, mp, upl, logger)
	svc.Start()
	defer svc.Stop()

	jobID := "test-job-info-err"
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
	mp.GetInfoMock.Set(func(_ context.Context, fp string) (*service.MediaInfo, error) {
		if fp != resultPath {
			t.Errorf("GetInfo called on unexpected file: %s", fp)
		}
		return &service.MediaInfo{Duration: 270 * time.Second, FileLenBytes: 3072}, nil
	})
	mp.ConcatenateMock.Set(func(_ context.Context, fps []string, codec string) (string, error) {
		return resultPath, nil
	})
	mp.ConcatDurationsMock.Set(func(_ context.Context, fps []string) ([]time.Duration, error) {
		return nil, errors.New("ffprobe: failed to read packets")
	})

	// AddChapterTags should NOT be called when ConcatDurations fails.
	addChapterTagsCalled := false
	mp.AddChapterTagsMock.Optional().Set(func(_ context.Context, fp string, chapters []service.Chapter) error {
		addChapterTagsCalled = true
		return nil
	})

	upl.UploadMock.Set(func(_ context.Context, fp string, url string) error {
		return nil
	})

	payload, _ := json.Marshal(jobID)
	if err := onJob(context.Background(), payload); err != nil {
		t.Fatalf("onJob failed: %v", err)
	}

	if addChapterTagsCalled {
		t.Error("AddChapterTags should not be called when ConcatDurations fails")
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
	var onJob func(ctx context.Context, payloadBytes []byte) error
	queue.SubscribeMock.Set(func(_ context.Context, _ string, f func(context.Context, []byte) error, _ func(context.Context, *sql.Tx, []byte, service.Attempt, error) error) {
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

	// ConcatDurations returns the durations of the files in the result. The
	// chapters must use them, not the durations of the files themselves.
	concatDurations := map[string]time.Duration{
		"/tmp/dl/intro.mp3":    1 * time.Minute,
		"/tmp/dl/chapter1.mp3": 2 * time.Minute,
		"/tmp/dl/chapter2.mp3": 90 * time.Second,
	}
	resultPath := "/tmp/result/output.mp3"
	mp.GetInfoMock.Set(func(_ context.Context, fp string) (*service.MediaInfo, error) {
		if fp != resultPath {
			t.Errorf("GetInfo called on unexpected file: %s", fp)
		}
		return &service.MediaInfo{Duration: 270 * time.Second, FileLenBytes: 3072}, nil
	})

	mp.ConcatenateMock.Set(func(_ context.Context, fps []string, codec string) (string, error) {
		return resultPath, nil
	})
	mp.ConcatDurationsMock.Set(func(_ context.Context, fps []string) ([]time.Duration, error) {
		durations := make([]time.Duration, len(fps))
		for i, fp := range fps {
			durations[i] = concatDurations[fp]
		}
		return durations, nil
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
	if err := onJob(context.Background(), payload); err != nil {
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
