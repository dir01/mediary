package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SQLiteBusyTimeoutMS is the busy_timeout (in milliseconds) applied to every
// SQLite connection. It must be large enough to absorb contention with
// litestream's WAL checkpointing and the jobs-queue worker's read transactions;
// without it, concurrent writes fail immediately with SQLITE_BUSY.
const SQLiteBusyTimeoutMS = 5000

// OpenSQLiteDB opens the SQLite database at path with the pragmas mediary relies
// on (WAL journal mode + busy_timeout) and verifies they actually took effect.
//
// The verification is deliberate: modernc.org/sqlite only honors the `_pragma`
// DSN form and silently ignores the mattn-style `_journal_mode`/`_busy_timeout`
// keys. A wrong DSN therefore leaves busy_timeout at 0 and surfaces much later
// as intermittent "database is locked (SQLITE_BUSY)" errors under load. Failing
// fast at open turns that into an immediate, obvious startup error and keeps
// production and tests from silently drifting apart.
func OpenSQLiteDB(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(%d)&_pragma=journal_mode(WAL)", path, SQLiteBusyTimeoutMS)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database: %w", err)
	}
	if err := verifySQLitePragmas(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// verifySQLitePragmas confirms the pragmas set via the DSN actually took effect.
func verifySQLitePragmas(db *sql.DB) error {
	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		return fmt.Errorf("reading journal_mode pragma: %w", err)
	}
	if journalMode != "wal" {
		return fmt.Errorf("expected WAL journal mode, got %q (DSN pragmas not applied?)", journalMode)
	}

	var busyTimeout int
	if err := db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		return fmt.Errorf("reading busy_timeout pragma: %w", err)
	}
	if busyTimeout != SQLiteBusyTimeoutMS {
		return fmt.Errorf("expected busy_timeout=%d, got %d (DSN pragmas not applied?)", SQLiteBusyTimeoutMS, busyTimeout)
	}
	return nil
}
