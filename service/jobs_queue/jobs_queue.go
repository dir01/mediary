package jobsqueue

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/dir01/sqlq"
	"github.com/samber/oops"
)

type SQLQ struct {
	queue  sqlq.JobsQueue
	logger *slog.Logger
}

func NewSQLJobsQueue(db *sql.DB, logger *slog.Logger) (*SQLQ, error) {
	// This is the deadline for the whole job and is the parent of every per-stage
	// context in the flows (download 1h + concat 1h + upload 2h). It must exceed the
	// sum of those stage budgets, otherwise time spent downloading/concatenating eats
	// into the upload's budget and a large S3 PUT gets cancelled before its own
	// per-stage timeout. 5h leaves headroom above the 4h worst-case sum.
	q, err := sqlq.New(db, sqlq.DBTypeSQLite,
		sqlq.WithDefaultJobTimeout(5*time.Hour),
	)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to create sqlq")
	}
	return &SQLQ{queue: q, logger: logger}, nil
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
