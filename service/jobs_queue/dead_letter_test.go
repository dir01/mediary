package jobsqueue

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/service/mocks"
	"github.com/dir01/mediary/storage"
	"github.com/gojuno/minimock/v3"
)

// TestDeadLetter_WorkerStoppedDuringLastAttempt examines a job whose worker process
// stopped (for example, an OOM kill) during each attempt. No flow returned, so no
// attempt wrote a result. When the queue takes the job again with no attempts left,
// it does not run the handler, and the dead letter hook must set the job to failed.
func TestDeadLetter_WorkerStoppedDuringLastAttempt(t *testing.T) {
	for _, tc := range []struct {
		name         string
		status       string
		expectStatus string
	}{
		{name: "stuck job becomes failed", status: service.JobStatusProcessing, expectStatus: service.JobStatusFailed},
		{name: "cancelled job stays cancelled", status: service.JobStatusCancelled, expectStatus: service.JobStatusCancelled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			// A file database, so that this test does not share the in-memory database of other tests.
			db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "test.db")+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
			if err != nil {
				t.Fatalf("error opening sqlite db: %v", err)
			}
			t.Cleanup(func() { _ = db.Close() })
			queue, err := NewSQLJobsQueue(db, logger)
			if err != nil {
				t.Fatalf("error creating sql job queue: %v", err)
			}
			t.Cleanup(queue.Shutdown)
			store, err := storage.NewSQLiteStorage(db)
			if err != nil {
				t.Fatalf("error creating storage: %v", err)
			}

			const jobID = "stuck-job"
			if err := store.SaveJob(ctx, &service.Job{
				JobParams:     service.JobParams{URL: "magnet:?xt=urn:btih:deadbeef", Type: "concatenate"},
				ID:            jobID,
				DisplayStatus: tc.status,
				Attempt:       6,
				MaxAttempts:   6,
			}); err != nil {
				t.Fatalf("error saving job: %v", err)
			}
			// Make the state that 6 stopped attempts leave: each attempt start added 1 to
			// retry_count, and no attempt wrote a result.
			if err := queue.Publish(ctx, "process", jobID); err != nil {
				t.Fatalf("error publishing job: %v", err)
			}
			if _, err := db.ExecContext(ctx, `UPDATE sqlq_jobs SET retry_count = 6, last_error = 'signal: killed'`); err != nil {
				t.Fatalf("error changing queue state: %v", err)
			}

			// The mocks have no expectations: the handler must not run.
			mc := minimock.NewController(t)
			svc := service.NewService(mocks.NewDownloaderMock(mc), store, queue,
				mocks.NewMediaProcessorMock(mc), mocks.NewUploaderMock(mc), logger)
			svc.Start()

			var job *service.Job
			if !eventually(20*time.Second, func() bool {
				var count int
				if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlq_dead_letter_queue`).Scan(&count); err != nil || count != 1 {
					return false
				}
				job, err = store.GetJob(ctx, jobID)
				return err == nil
			}) {
				t.Fatal("job was not moved to the dead letter queue")
			}
			if job.DisplayStatus != tc.expectStatus {
				data, _ := json.Marshal(job)
				t.Fatalf("status = %q, want %q: %s", job.DisplayStatus, tc.expectStatus, data)
			}
			if tc.expectStatus == service.JobStatusFailed {
				if !strings.Contains(job.Error, "no attempts left") || !strings.Contains(job.Error, "signal: killed") {
					t.Errorf("unexpected error: %q", job.Error)
				}
				if job.Attempt != 6 || job.MaxAttempts != 6 {
					t.Errorf("attempt = %d of %d, want 6 of 6", job.Attempt, job.MaxAttempts)
				}
			}
		})
	}
}
