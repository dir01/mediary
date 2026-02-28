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

	_ "modernc.org/sqlite"
)

var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

func TestSQLJobsQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

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
		})

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
		})

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
	db, err := sql.Open("sqlite", "file::memory:?cache=shared&_journal_mode=WAL")
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
