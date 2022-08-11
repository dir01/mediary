package mediary

import (
	"context"
)

func NewStorageInMemory() Storage {
	return &StorageMemory{metadata: make(map[string]*Metadata)}
}

type StorageMemory struct {
	metadata map[string]*Metadata
}

func (s *StorageMemory) GetMetadata(ctx context.Context, url string) (*Metadata, error) {
	return s.metadata[url], nil
}

func (s *StorageMemory) SaveMetadata(ctx context.Context, url string, metadata *Metadata) error {
	s.metadata[url] = metadata
	return nil
}
