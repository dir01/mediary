package jobsqueue

import (
	"context"
	"database/sql"
	"log/slog"
	"math/rand"
	"time"

	"github.com/dir01/sqlq"
	"github.com/samber/oops"
)

type SQLQ struct {
	queue  sqlq.JobsQueue
	logger *slog.Logger
}

func NewSQLJobsQueue(db *sql.DB, logger *slog.Logger) (*SQLQ, error) {
	// 2h is a hard total budget for a whole job, and the parent of every per-stage
	// context in the flows. Each stage ceiling (download, concat, upload) is therefore
	// capped by whatever time remains in this budget; the stage ceilings are kept well
	// under 2h so a slow download/concat can't starve the upload.
	q, err := sqlq.New(db, sqlq.DBTypeSQLite,
		sqlq.WithDefaultJobTimeout(2*time.Hour),
		// Default is 3 retries; every attempt re-downloads and re-concatenates, so
		// retries are expensive — keep the count modest.
		sqlq.WithDefaultMaxRetries(5),
		// Default backoff is an uncapped 2^n seconds. Use a larger base with a cap so
		// transient failures get real spacing without ever scheduling a retry that is
		// hours/days out.
		sqlq.WithDefaultBackoffFunc(retryBackoff),
	)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to create sqlq")
	}
	return &SQLQ{queue: q, logger: logger}, nil
}

// retryBackoff is exponential with jitter, capped: ~10s, 20s, 40s, 80s, 160s, …, 30m.
// The first retry stays quick so a transient failure recovers fast, then spacing grows
// and is capped at 30m so a persistently-failing job never schedules a retry hours out.
// sqlq calls this with retryNum = RetryCount+1 (1 for the first retry).
func retryBackoff(retryNum uint16) time.Duration {
	const base = 10 * time.Second
	const maxBackoff = 30 * time.Minute
	d := base
	for i := uint16(1); i < retryNum && d < maxBackoff; i++ {
		d *= 2
	}
	if d > maxBackoff {
		d = maxBackoff
	}
	// jitter: up to 20%, to avoid retries from many jobs aligning
	return d + time.Duration(rand.Int63n(int64(d)/5+1))
}

func (s *SQLQ) Run() {
	s.queue.Run()
}

func (s *SQLQ) Shutdown() {
	s.queue.Shutdown()
}

func (s *SQLQ) Publish(ctx context.Context, jobType string, payload any) error {
	if err := s.queue.Publish(ctx, jobType, payload); err != nil {
		return oops.Wrapf(err, "failed to publish job")
	}
	return nil
}

func (s *SQLQ) Subscribe(ctx context.Context, jobType string, f func(ctx context.Context, payloadBytes []byte) error) {
	err := s.queue.Consume(ctx, jobType, func(ctx context.Context, tx *sql.Tx, payloadBytes []byte) error {
		return f(ctx, payloadBytes)
	})
	if err != nil {
		s.logger.Error("failed to register consumer", slog.String("jobType", jobType), slog.Any("error", err))
	}
}
