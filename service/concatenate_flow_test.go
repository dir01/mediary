package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/service/mocks"
	"github.com/gojuno/minimock/v3"
)

// TestConcatenateFlow_GetInfoErrorSkipsChapters verifies that when GetInfo
// returns an error for any file, chapter tags are skipped but concatenation
// still succeeds. Before the fix, GetDuration used zaperr.Wrap(nil, ...)
// which silently returned (0, nil), causing chapters with all-zero timestamps.
// Now GetDuration uses ffprobe and properly reports errors.
func TestConcatenateFlow_GetInfoErrorSkipsChapters(t *testing.T) {
	mc := minimock.NewController(t)

	storage := mocks.NewStorageMock(mc)
	queue := mocks.NewJobsQueueMock(mc)
	dwn := mocks.NewDownloaderMock(mc)
	mp := mocks.NewMediaProcessorMock(mc)
	upl := mocks.NewUploaderMock(mc)

	var onJob func(ctx context.Context, payloadBytes []byte) error
	queue.SubscribeMock.Set(func(_ context.Context, _ string, f func(context.Context, []byte) error) {
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
	getInfoCalls := 0

	// GetInfo fails on the second file (simulates ffprobe parse error).
	// The first call (for individual file) returns an error;
	// the last call (for the concatenated result) must succeed.
	mp.GetInfoMock.Set(func(_ context.Context, fp string) (*service.MediaInfo, error) {
		getInfoCalls++
		if fp == resultPath {
			return &service.MediaInfo{Duration: 270 * time.Second, FileLenBytes: 3072}, nil
		}
		return nil, errors.New("ffprobe: failed to parse duration")
	})

	mp.ConcatenateMock.Set(func(_ context.Context, fps []string, codec string) (string, error) {
		return resultPath, nil
	})

	// AddChapterTags should NOT be called when GetInfo fails.
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
		t.Error("AddChapterTags should not be called when GetInfo fails")
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
	queue.SubscribeMock.Set(func(_ context.Context, _ string, f func(context.Context, []byte) error) {
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
