package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/dir01/mediary/service"
)

func NewRedisStorage(rediClient *redis.Client, keyPrefix string) service.Storage {
	return &RedisStorage{
		redisClient: rediClient,
		keyPrefix:   keyPrefix,
	}
}

type RedisStorage struct {
	redisClient *redis.Client
	keyPrefix   string
}

func (s *RedisStorage) GetJob(ctx context.Context, id string) (*service.Job, error) {
	ctx, span := otel.Tracer("github.com/dir01/mediary/storage").Start(ctx, "storage.GetJob",
		trace.WithAttributes(attribute.String("job.id", id)),
	)
	defer span.End()

	jobBytes, err := s.redisClient.Get(ctx, s.jobKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	job := &service.Job{}
	if err := json.Unmarshal(jobBytes, job); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetAttributes(attribute.String("job.status", job.DisplayStatus))
	return job, nil
}

func (s *RedisStorage) SaveJob(ctx context.Context, job *service.Job) error {
	ctx, span := otel.Tracer("github.com/dir01/mediary/storage").Start(ctx, "storage.SaveJob",
		trace.WithAttributes(
			attribute.String("job.id", job.ID),
			attribute.String("job.status", job.DisplayStatus),
		),
	)
	defer span.End()

	jobBytes, err := json.Marshal(job)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if err := s.redisClient.Set(ctx, s.jobKey(job.ID), jobBytes, 0).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (s *RedisStorage) GetMetadata(ctx context.Context, url string) (*service.Metadata, error) {
	ctx, span := otel.Tracer("github.com/dir01/mediary/storage").Start(ctx, "storage.GetMetadata",
		trace.WithAttributes(attribute.String("url", url)),
	)
	defer span.End()

	metaBytes, err := s.redisClient.Get(ctx, s.metadataKey(url)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	meta := &service.Metadata{}
	if err := json.Unmarshal(metaBytes, meta); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return meta, nil
}

func (s *RedisStorage) SaveMetadata(ctx context.Context, metadata *service.Metadata) error {
	ctx, span := otel.Tracer("github.com/dir01/mediary/storage").Start(ctx, "storage.SaveMetadata",
		trace.WithAttributes(attribute.String("url", metadata.URL)),
	)
	defer span.End()

	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if err := s.redisClient.Set(ctx, s.metadataKey(metadata.URL), metaBytes, 0).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (s *RedisStorage) jobKey(id string) string {
	return fmt.Sprintf("%s:job:%s", s.keyPrefix, id)
}

func (s *RedisStorage) metadataKey(url string) string {
	return fmt.Sprintf("%s:metadata:%s", s.keyPrefix, url)
}
