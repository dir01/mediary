package storage

import (
	"context"

	"github.com/dir01/mediary/service"
)

func NewStorageInMemory() service.Storage {
	return &StorageMemory{
		metadataMap: make(map[string]*service.Metadata),
		jobMap:      make(map[string]*service.JobState),
	}
}

type StorageMemory struct {
	metadataMap map[string]*service.Metadata
	jobMap      map[string]*service.JobState
}

func (s *StorageMemory) GetJob(ctx context.Context, id string) (*service.JobState, error) {
	return s.jobMap[id], nil
}

func (s *StorageMemory) SaveJob(ctx context.Context, job *service.JobState) error {
	s.jobMap[job.ID] = job
	return nil
}

func (s *StorageMemory) GetMetadata(ctx context.Context, url string) (*service.Metadata, error) {
	return s.metadataMap[url], nil
}

func (s *StorageMemory) SaveMetadata(ctx context.Context, metadata *service.Metadata) error {
	s.metadataMap[metadata.URL] = metadata
	return nil
}
