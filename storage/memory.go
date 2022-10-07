package storage

import (
	"context"
	"sync"

	"github.com/dir01/mediary/service"
)

func NewMemoryStorage() service.Storage {
	return &MemoryStorage{
		metadataMap: make(map[string]service.Metadata),
		jobMap:      make(map[string]service.Job),
	}
}

type MemoryStorage struct {
	metadataMap   map[string]service.Metadata
	metadataMutex sync.RWMutex
	jobMap        map[string]service.Job
	jobMutex      sync.RWMutex
}

func (s *MemoryStorage) GetJob(ctx context.Context, id string) (*service.Job, error) {
	s.jobMutex.RLock()
	defer s.jobMutex.RUnlock()
	if job, exists := s.jobMap[id]; exists {
		return &job, nil
	} else {
		return nil, nil
	}
}

func (s *MemoryStorage) SaveJob(ctx context.Context, job *service.Job) error {
	if job == nil {
		return nil
	}
	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()
	s.jobMap[job.ID] = *job
	return nil
}

func (s *MemoryStorage) GetMetadata(ctx context.Context, url string) (*service.Metadata, error) {
	s.metadataMutex.RLock()
	defer s.metadataMutex.RUnlock()

	if md, exists := s.metadataMap[url]; exists {
		return &md, nil
	} else {
		return nil, nil
	}
}

func (s *MemoryStorage) SaveMetadata(ctx context.Context, metadata *service.Metadata) error {
	if metadata == nil {
		return nil
	}
	s.metadataMutex.Lock()
	defer s.metadataMutex.Unlock()
	s.metadataMap[metadata.URL] = *metadata
	return nil
}
