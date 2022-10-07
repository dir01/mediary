package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dir01/mediary/service"
	"github.com/go-redis/redis"
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
	jobBytes, err := s.redisClient.Get(s.jobKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	job := &service.Job{}
	if err := json.Unmarshal(jobBytes, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *RedisStorage) SaveJob(ctx context.Context, job *service.Job) error {
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return s.redisClient.Set(s.jobKey(job.ID), jobBytes, 0).Err()
}

func (s *RedisStorage) GetMetadata(ctx context.Context, url string) (*service.Metadata, error) {
	metaBytes, err := s.redisClient.Get(s.metadataKey(url)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	meta := &service.Metadata{}
	if err := json.Unmarshal(metaBytes, meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func (s *RedisStorage) SaveMetadata(ctx context.Context, metadata *service.Metadata) error {
	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	return s.redisClient.Set(s.metadataKey(metadata.URL), metaBytes, 0).Err()
}

func (s *RedisStorage) jobKey(id string) string {
	return fmt.Sprintf("%s:job:%s", s.keyPrefix, id)
}

func (s *RedisStorage) metadataKey(url string) string {
	return fmt.Sprintf("%s:metadata:%s", s.keyPrefix, url)
}
