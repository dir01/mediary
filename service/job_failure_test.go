package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/service/mocks"
	"github.com/gojuno/minimock/v3"
)

// TestFailedAttemptIsRecorded verifies that a failed attempt writes the error and
// the attempt to the job, and that the status is "failed" only after the final attempt.
func TestFailedAttemptIsRecorded(t *testing.T) {
	const jobID = "failing-job"
	for _, tc := range []struct {
		name         string
		attempt      service.Attempt
		expectStatus string
	}{
		{
			name:         "non-final attempt is retrying",
			attempt:      service.Attempt{Number: 2, Max: 6},
			expectStatus: service.JobStatusRetrying,
		},
		{
			name:         "final attempt is failed",
			attempt:      service.Attempt{Number: 6, Max: 6, IsFinal: true},
			expectStatus: service.JobStatusFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mc := minimock.NewController(t)
			dwn := mocks.NewDownloaderMock(mc)
			storage := mocks.NewStorageMock(mc)
			queue := mocks.NewJobsQueueMock(mc)

			var handler func(context.Context, []byte) error
			queue.
				RunMock.Optional().Set(func() {}).
				ShutdownMock.Optional().Set(func() {}).
				SubscribeMock.Set(func(ctx context.Context, jobType string, f func(context.Context, []byte) error) {
				handler = f
			})

			// Storage keeps one stored copy of the job.
			var mu sync.Mutex
			stored := service.Job{
				JobParams: service.JobParams{
					URL:  "magnet:?xt=urn:btih:deadbeef",
					Type: "concatenate",
					Params: map[string]interface{}{
						"variants":  []interface{}{"01.mp3", "02.mp3"},
						"uploadUrl": "http://example.invalid/upload",
					},
				},
				ID:            jobID,
				DisplayStatus: service.JobStatusCreated,
			}
			storage.GetJobMock.Set(func(ctx context.Context, id string) (*service.Job, error) {
				mu.Lock()
				defer mu.Unlock()
				j := stored
				return &j, nil
			})
			storage.SaveJobMock.Set(func(ctx context.Context, job *service.Job) error {
				mu.Lock()
				defer mu.Unlock()
				stored = *job
				return nil
			})
			dwn.DownloadMock.Set(func(ctx context.Context, url string, filepaths []string) (map[string]string, error) {
				return nil, errors.New("download timed out: 1 of 2 files complete (3 MB of 7 MB)")
			})

			svc := service.NewService(dwn, storage, queue, nil, nil, logger)
			svc.Start()
			defer svc.Stop()

			payload, _ := json.Marshal(jobID)
			ctx := service.ContextWithAttempt(context.Background(), tc.attempt)
			if err := handler(ctx, payload); err == nil {
				t.Fatal("expected the handler to return the flow error")
			}

			if stored.DisplayStatus != tc.expectStatus {
				t.Errorf("status = %q, want %q", stored.DisplayStatus, tc.expectStatus)
			}
			wantErr := "failed to download variants: download timed out: 1 of 2 files complete (3 MB of 7 MB)"
			if stored.Error != wantErr {
				t.Errorf("error = %q, want %q", stored.Error, wantErr)
			}
			if stored.Attempt != tc.attempt.Number || stored.MaxAttempts != tc.attempt.Max {
				t.Errorf("attempt = %d of %d, want %d of %d",
					stored.Attempt, stored.MaxAttempts, tc.attempt.Number, tc.attempt.Max)
			}
		})
	}
}
