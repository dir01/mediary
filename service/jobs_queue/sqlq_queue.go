package jobsqueue

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/dir01/sqlq"
	"github.com/samber/oops"
)

type SQLQ struct {
	queue  sqlq.JobsQueue
	logger *slog.Logger
}

func NewSQLJobsQueue(db *sql.DB, logger *slog.Logger) (*SQLQ, error) {
	q, err := sqlq.New(db, sqlq.DBTypeSQLite)
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

func (s *SQLQ) Subscribe(ctx context.Context, jobType string, f func(payloadBytes []byte) error) {
	err := s.queue.Consume(ctx, jobType, func(ctx context.Context, tx *sql.Tx, payloadBytes []byte) error {
		return f(payloadBytes)
	})
	if err != nil {
		s.logger.Error("failed to register consumer", slog.String("jobType", jobType), slog.Any("error", err))
	}
}

