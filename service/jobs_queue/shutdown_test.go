package jobsqueue

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/dir01/mediary/service"
)

// TestShutdown_ReleasesClaimOfRunningJob verifies that Shutdown stops a running
// handler and releases the claim of its job. Then the next process takes the job
// at once, and does not wait for the claim to expire (the job timeout, 5h).
func TestShutdown_ReleasesClaimOfRunningJob(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	openDB := func() *sql.DB {
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			t.Fatalf("error opening sqlite db: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })
		return db
	}

	// The first process starts the job, and then stops.
	first, err := NewSQLJobsQueue(openDB(), logger)
	if err != nil {
		t.Fatalf("error creating sql job queue: %v", err)
	}
	started := make(chan struct{})
	stopped := make(chan error, 1)
	first.Subscribe(ctx, "process", func(ctx context.Context, _ []byte) error {
		close(started)
		<-ctx.Done() // A long download, for example.
		stopped <- ctx.Err()
		return ctx.Err()
	}, nil)
	if err := first.Publish(ctx, "process", "job-1"); err != nil {
		t.Fatalf("error publishing job: %v", err)
	}
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("job was never started")
	}
	first.Shutdown()
	select {
	case err := <-stopped:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("handler stopped with %v, want context.Canceled", err)
		}
	default:
		t.Fatal("Shutdown did not stop the running handler")
	}

	// The next process must get the same job at once, as attempt 1 again.
	next, err := NewSQLJobsQueue(openDB(), logger)
	if err != nil {
		t.Fatalf("error creating sql job queue: %v", err)
	}
	t.Cleanup(next.Shutdown)
	attempts := make(chan service.Attempt, 1)
	next.Subscribe(ctx, "process", func(ctx context.Context, _ []byte) error {
		attempts <- service.AttemptFromContext(ctx)
		return nil
	}, nil)
	select {
	case got := <-attempts:
		if got.Number != 1 {
			t.Errorf("attempt = %d, want 1: an attempt that the shutdown stopped must not count", got.Number)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("the next process did not get the job: its claim was not released")
	}
}
