package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/service/mocks"
	"github.com/gojuno/minimock/v3"
)

func TestDeleteJob(t *testing.T) {
	const id = "job-123"
	existing := &service.Job{JobParams: service.JobParams{}, ID: id, DisplayStatus: service.JobStatusCancelled}

	for _, tc := range []struct {
		name           string
		getJob         *service.Job
		getErr         error
		deleteErr      error
		expectDelete   bool
		expectJob      *service.Job
		expectErrMatch string
	}{
		{
			name:         "existing job is hard-deleted and returned",
			getJob:       existing,
			expectDelete: true,
			expectJob:    existing,
		},
		{
			name:         "missing job returns nil and is not deleted",
			getJob:       nil,
			expectDelete: false,
			expectJob:    nil,
		},
		{
			name:           "get error is propagated without deleting",
			getErr:         errors.New("boom"),
			expectDelete:   false,
			expectErrMatch: "boom",
		},
		{
			name:           "delete error is wrapped",
			getJob:         existing,
			deleteErr:      errors.New("db down"),
			expectDelete:   true,
			expectErrMatch: "failed to delete job",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mc := minimock.NewController(t)
			dwn := mocks.NewDownloaderMock(mc)
			storage := mocks.NewStorageMock(mc)
			queue := mocks.NewJobsQueueMock(mc)
			queue.
				SubscribeMock.Optional().Set(func(ctx context.Context, jobType string, f1 func(context.Context, []byte) error) {}).
				RunMock.Optional().Set(func() {}).
				PublishMock.Optional().Set(func(ctx context.Context, jobType string, payload any) error { return nil }).
				ShutdownMock.Optional().Set(func() {})

			svc := service.NewService(dwn, storage, queue, nil, nil, logger)
			svc.Start()
			defer svc.Stop()

			storage.GetJobMock.Set(func(ctx context.Context, gotID string) (*service.Job, error) {
				if gotID != id {
					t.Fatalf("GetJob called with %q, want %q", gotID, id)
				}
				return tc.getJob, tc.getErr
			})

			deleteCalled := false
			storage.DeleteJobMock.Optional().Set(func(ctx context.Context, gotID string) error {
				deleteCalled = true
				if gotID != id {
					t.Fatalf("DeleteJob called with %q, want %q", gotID, id)
				}
				return tc.deleteErr
			})

			job, err := svc.DeleteJob(context.TODO(), id)

			if deleteCalled != tc.expectDelete {
				t.Errorf("DeleteJob called = %v, want %v", deleteCalled, tc.expectDelete)
			}
			if tc.expectErrMatch != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectErrMatch) {
					t.Errorf("expected error containing %q, got %v", tc.expectErrMatch, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if job != tc.expectJob {
				t.Errorf("expected job %v, got %v", tc.expectJob, job)
			}
		})
	}
}

// TestCancelStopsRunningFlow verifies that CancelJob interrupts a flow that is
// mid-execution (blocked in the download stage), not just one that is queued.
func TestCancelStopsRunningFlow(t *testing.T) {
	const jobID = "running-job"
	makeJob := func() *service.Job {
		return &service.Job{
			JobParams: service.JobParams{
				URL:  "magnet:?xt=urn:btih:deadbeef",
				Type: "concatenate",
				Params: map[string]interface{}{
					"variants":  []interface{}{"file.mp3"},
					"uploadUrl": "http://example.invalid/upload",
				},
			},
			ID:            jobID,
			DisplayStatus: service.JobStatusCreated,
		}
	}

	mc := minimock.NewController(t)
	dwn := mocks.NewDownloaderMock(mc)
	storage := mocks.NewStorageMock(mc)
	queue := mocks.NewJobsQueueMock(mc)

	var handler func(context.Context, []byte) error
	queue.
		RunMock.Optional().Set(func() {}).
		ShutdownMock.Optional().Set(func() {}).
		PublishMock.Optional().Set(func(ctx context.Context, jobType string, payload any) error { return nil }).
		SubscribeMock.Set(func(ctx context.Context, jobType string, f func(context.Context, []byte) error) {
			handler = f
		})

	storage.GetJobMock.Set(func(ctx context.Context, id string) (*service.Job, error) { return makeJob(), nil })
	storage.SaveJobMock.Set(func(ctx context.Context, job *service.Job) error { return nil })

	downloadStarted := make(chan struct{})
	dwn.DownloadMock.Set(func(ctx context.Context, url string, filepaths []string) (map[string]string, error) {
		close(downloadStarted)
		<-ctx.Done() // block until the flow's context is cancelled
		return nil, ctx.Err()
	})

	svc := service.NewService(dwn, storage, queue, nil, nil, logger)
	svc.Start()
	defer svc.Stop()
	if handler == nil {
		t.Fatal("Start did not subscribe a handler")
	}

	payload, err := json.Marshal(jobID)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- handler(context.Background(), payload) }()

	select {
	case <-downloadStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("flow never reached the download stage")
	}

	if _, err := svc.CancelJob(context.Background(), jobID); err != nil {
		t.Fatalf("CancelJob: %v", err)
	}

	select {
	case <-done: // flow unwound after cancellation — success
	case <-time.After(3 * time.Second):
		t.Fatal("flow did not stop after CancelJob")
	}
}

// TestCancelledJobSkipsExecution verifies that a job cancelled while still queued
// is not executed when the worker later picks it up (P2 regression).
func TestCancelledJobSkipsExecution(t *testing.T) {
	const jobID = "cancelled-queued-job"
	mc := minimock.NewController(t)
	dwn := mocks.NewDownloaderMock(mc) // Download intentionally left unconfigured
	storage := mocks.NewStorageMock(mc)
	queue := mocks.NewJobsQueueMock(mc)

	var handler func(context.Context, []byte) error
	queue.
		RunMock.Optional().Set(func() {}).
		ShutdownMock.Optional().Set(func() {}).
		PublishMock.Optional().Set(func(ctx context.Context, jobType string, payload any) error { return nil }).
		SubscribeMock.Set(func(ctx context.Context, jobType string, f func(context.Context, []byte) error) {
			handler = f
		})

	storage.GetJobMock.Set(func(ctx context.Context, id string) (*service.Job, error) {
		return &service.Job{
			JobParams: service.JobParams{
				URL:  "magnet:?xt=urn:btih:deadbeef",
				Type: "concatenate",
				Params: map[string]interface{}{
					"variants":  []interface{}{"file.mp3"},
					"uploadUrl": "http://example.invalid/upload",
				},
			},
			ID:            jobID,
			DisplayStatus: service.JobStatusCancelled,
		}, nil
	})

	svc := service.NewService(dwn, storage, queue, nil, nil, logger)
	svc.Start()
	defer svc.Stop()

	payload, err := json.Marshal(jobID)
	if err != nil {
		t.Fatal(err)
	}
	// If the flow ran, it would call the unconfigured Download mock and panic.
	if err := handler(context.Background(), payload); err != nil {
		t.Fatalf("handler returned error for cancelled job: %v", err)
	}
}

// TestCreateJobRefreshesUploadURLOnDuplicate verifies that a duplicate submit with
// a fresh presigned upload URL updates the stored job (P1 regression).
func TestCreateJobRefreshesUploadURLOnDuplicate(t *testing.T) {
	const oldURL = "http://s3.local/bucket/key?X-Amz-Signature=OLD"
	const newURL = "http://s3.local/bucket/key?X-Amz-Signature=NEW"

	mc := minimock.NewController(t)
	dwn := mocks.NewDownloaderMock(mc)
	storage := mocks.NewStorageMock(mc)
	queue := mocks.NewJobsQueueMock(mc)
	queue.PublishMock.Optional().Set(func(ctx context.Context, jobType string, payload any) error { return nil })

	existing := &service.Job{
		JobParams: service.JobParams{
			URL:  "magnet:?xt=urn:btih:deadbeef",
			Type: "concatenate",
			Params: map[string]interface{}{
				"variants":  []interface{}{"file.mp3"},
				"uploadUrl": oldURL,
			},
		},
		ID:            "existing-id",
		DisplayStatus: service.JobStatusDownloading,
	}
	storage.GetJobMock.Set(func(ctx context.Context, id string) (*service.Job, error) { return existing, nil })

	var savedURL string
	storage.SaveJobMock.Set(func(ctx context.Context, job *service.Job) error {
		savedURL, _ = job.Params["uploadUrl"].(string)
		return nil
	})

	svc := service.NewService(dwn, storage, queue, nil, nil, logger)

	params := &service.JobParams{
		URL:  "magnet:?xt=urn:btih:deadbeef",
		Type: "concatenate",
		Params: map[string]interface{}{
			"variants":  []interface{}{"file.mp3"},
			"uploadUrl": newURL,
		},
	}
	job, err := svc.CreateJob(context.Background(), params)
	if !errors.Is(err, service.ErrJobAlreadyExists) {
		t.Fatalf("expected ErrJobAlreadyExists, got %v", err)
	}
	if got, _ := job.Params["uploadUrl"].(string); got != newURL {
		t.Errorf("returned job upload url = %q, want refreshed %q", got, newURL)
	}
	if savedURL != newURL {
		t.Errorf("stored job upload url = %q, want refreshed %q", savedURL, newURL)
	}
}
