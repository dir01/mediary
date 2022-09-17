package storage

import (
	"context"
	"sync"

	"github.com/dir01/mediary/service"
)

func NewStorageInMemory() service.Storage {
	return &StorageMemory{
		metadataMap: make(map[string]service.Metadata),
		jobMap:      make(map[string]service.JobState),
	}
}

type StorageMemory struct {
	metadataMap   map[string]service.Metadata
	metadataMutex sync.RWMutex
	jobMap        map[string]service.JobState
	jobMutex      sync.RWMutex
}

func (s *StorageMemory) GetJob(ctx context.Context, id string) (*service.JobState, error) {
	s.jobMutex.RLock()
	defer s.jobMutex.RUnlock()
	if job, exists := s.jobMap[id]; exists {
		return &job, nil
	} else {
		return nil, nil
	}
}

func (s *StorageMemory) SaveJob(ctx context.Context, job *service.JobState) error {
	if job == nil {
		return nil
	}
	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()
	s.jobMap[job.ID] = *job
	return nil
}

func (s *StorageMemory) GetMetadata(ctx context.Context, url string) (*service.Metadata, error) {
	s.metadataMutex.RLock()
	defer s.metadataMutex.RUnlock()

	if md, exists := s.metadataMap[url]; exists {
		return &md, nil
	} else {
		return nil, nil
	}
}

func (s *StorageMemory) SaveMetadata(ctx context.Context, metadata *service.Metadata) error {
	if metadata == nil {
		return nil
	}
	s.metadataMutex.Lock()
	defer s.metadataMutex.Unlock()
	s.metadataMap[metadata.URL] = *metadata
	return nil
}
