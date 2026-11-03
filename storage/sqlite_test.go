package storage_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/storage"
	_ "modernc.org/sqlite"
)

// TestSQLiteStorage_KeepsRunID verifies that storage keeps the RunID of a job, and that
// the JSON of the job (the API response) does not show it.
func TestSQLiteStorage_KeepsRunID(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("error opening sqlite db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	store, err := storage.NewSQLiteStorage(db)
	if err != nil {
		t.Fatalf("error creating storage: %v", err)
	}

	job := &service.Job{ID: "job-1", DisplayStatus: service.JobStatusCreated, RunID: "run-1"}
	if err := store.SaveJob(ctx, job); err != nil {
		t.Fatalf("error saving job: %v", err)
	}
	got, err := store.GetJob(ctx, "job-1")
	if err != nil || got == nil || got.RunID != "run-1" {
		t.Fatalf("expected RunID %q, got %+v, %v", "run-1", got, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("error starting tx: %v", err)
	}
	if err := store.UpdateJobInTx(ctx, tx, "job-1", func(job *service.Job) bool {
		job.DisplayStatus = service.JobStatusFailed
		return true
	}); err != nil {
		t.Fatalf("error updating job: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("error committing tx: %v", err)
	}
	got, err = store.GetJob(ctx, "job-1")
	if err != nil || got.RunID != "run-1" || got.DisplayStatus != service.JobStatusFailed {
		t.Fatalf("UpdateJobInTx lost data: %+v, %v", got, err)
	}

	apiJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("error marshaling job: %v", err)
	}
	if strings.Contains(string(apiJSON), "run-1") {
		t.Errorf("API JSON shows the RunID: %s", apiJSON)
	}
}
