package jobsqueue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/dir01/mediary/service"
	_ "modernc.org/sqlite"
)

var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

func TestSQLJobsQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	t.Run("handler gets the attempt", func(t *testing.T) {
		db := openTestDB(t)
		queue, err := NewSQLJobsQueue(db, logger)
		if err != nil {
			t.Fatalf("error creating sql job queue: %v", err)
		}
		queue.Run()
		defer queue.Shutdown()

		attempts := make(chan service.Attempt, 1)
		queue.Subscribe(ctx, "attempt-job-type", func(ctx context.Context, _ []byte) error {
			attempts <- service.AttemptFromContext(ctx)
			return nil
		}, nil)
		if err := queue.Publish(ctx, "attempt-job-type", "x"); err != nil {
			t.Fatalf("error publishing job: %v", err)
		}

		select {
		case got := <-attempts:
			want := service.Attempt{Number: 1, Max: 6, IsFinal: false}
			if got != want {
				t.Errorf("attempt = %+v, want %+v", got, want)
			}
		case <-time.After(20 * time.Second):
			t.Fatal("job was never delivered to subscriber")
		}
	})

	t.Run("job is persisted", func(t *testing.T) {
		// First publish, then subscribe. Job should arrive.
		db := openTestDB(t)

		queue, err := NewSQLJobsQueue(db, logger)
		if err != nil {
			t.Fatalf("error creating sql job queue: %v", err)
		}
		queue.Run()
		defer queue.Shutdown()

		err = queue.Publish(ctx, "some-job-type", map[string]string{"foo": "bar"})
		if err != nil {
			t.Errorf("error publishing job: %v", err)
		}

		var callCountMutex sync.RWMutex
		callCount := 0
		queue.Subscribe(ctx, "some-job-type", func(_ context.Context, payloadBytes []byte) error {
			var result map[string]string
			err := json.Unmarshal(payloadBytes, &result)
			if err != nil {
				return err
			}
			callCountMutex.Lock()
			defer callCountMutex.Unlock()
			callCount++
			return nil
		}, nil)

		if eventually(20*time.Second, func() bool {
			callCountMutex.RLock()
			defer callCountMutex.RUnlock()
			return callCount == 1
		}) != true {
			t.Errorf("job was never delivered to subscriber")
		}
	})

	t.Run("job is retried", func(t *testing.T) {
		// If job fails once, it will be retried shortly after.
		db := openTestDB(t)

		queue, err := NewSQLJobsQueue(db, logger)
		if err != nil {
			t.Fatalf("error creating sql job queue: %v", err)
		}
		queue.Run()
		defer queue.Shutdown()

		err = queue.Publish(ctx, "some-job-type", map[string]string{"foo": "bar"})
		if err != nil {
			t.Errorf("error publishing job: %v", err)
		}

		var callCountMutex sync.RWMutex
		callCount := 0
		queue.Subscribe(ctx, "some-job-type", func(_ context.Context, payloadBytes []byte) error {
			callCountMutex.Lock()
			defer callCountMutex.Unlock()
			callCount++
			if callCount < 2 {
				return fmt.Errorf("some error")
			}
			return nil
		}, nil)

		if eventually(60*time.Second, func() bool {
			callCountMutex.RLock()
			defer callCountMutex.RUnlock()
			return callCount == 2
		}) != true {
			t.Errorf("job was never retried")
		}
	})
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file::memory:?cache=shared&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		t.Fatalf("error opening in-memory sqlite db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func eventually(timeout time.Duration, f func() bool) bool {
	timeoutChan := time.After(timeout)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-timeoutChan:
			return false
		case <-tick.C:
			if f() {
				return true
			}
		}
	}
}
