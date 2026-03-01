package storage

import (
	"context"
	"database/sql"
	"encoding/json"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/dir01/mediary/service"
)

func NewSQLiteStorage(db *sql.DB) (service.Storage, error) {
	s := &SQLiteStorage{db: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

type SQLiteStorage struct {
	db *sql.DB
}

func (s *SQLiteStorage) initSchema() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS mediary_jobs (
			id   TEXT PRIMARY KEY,
			data BLOB NOT NULL
		);
		CREATE TABLE IF NOT EXISTS mediary_metadata (
			url  TEXT PRIMARY KEY,
			data BLOB NOT NULL
		);
	`)
	return err
}

func (s *SQLiteStorage) GetJob(ctx context.Context, id string) (*service.Job, error) {
	ctx, span := otel.Tracer("github.com/dir01/mediary/storage").Start(ctx, "storage.GetJob",
		trace.WithAttributes(attribute.String("job.id", id)),
	)
	defer span.End()

	var data []byte
	err := s.db.QueryRowContext(ctx, `SELECT data FROM mediary_jobs WHERE id = ?`, id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	job := &service.Job{}
	if err := json.Unmarshal(data, job); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetAttributes(attribute.String("job.status", job.DisplayStatus))
	return job, nil
}

func (s *SQLiteStorage) SaveJob(ctx context.Context, job *service.Job) error {
	ctx, span := otel.Tracer("github.com/dir01/mediary/storage").Start(ctx, "storage.SaveJob",
		trace.WithAttributes(
			attribute.String("job.id", job.ID),
			attribute.String("job.status", job.DisplayStatus),
		),
	)
	defer span.End()

	data, err := json.Marshal(job)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT OR REPLACE INTO mediary_jobs (id, data) VALUES (?, ?)`, job.ID, data)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (s *SQLiteStorage) GetMetadata(ctx context.Context, url string) (*service.Metadata, error) {
	ctx, span := otel.Tracer("github.com/dir01/mediary/storage").Start(ctx, "storage.GetMetadata",
		trace.WithAttributes(attribute.String("url", url)),
	)
	defer span.End()

	var data []byte
	err := s.db.QueryRowContext(ctx, `SELECT data FROM mediary_metadata WHERE url = ?`, url).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	meta := &service.Metadata{}
	if err := json.Unmarshal(data, meta); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return meta, nil
}

func (s *SQLiteStorage) SaveMetadata(ctx context.Context, metadata *service.Metadata) error {
	ctx, span := otel.Tracer("github.com/dir01/mediary/storage").Start(ctx, "storage.SaveMetadata",
		trace.WithAttributes(attribute.String("url", metadata.URL)),
	)
	defer span.End()

	data, err := json.Marshal(metadata)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT OR REPLACE INTO mediary_metadata (url, data) VALUES (?, ?)`, metadata.URL, data)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}
