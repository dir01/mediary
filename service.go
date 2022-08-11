package mediary

import (
	"context"
	"errors"
)

func NewService(downloaders []Downloader, storage Storage) *Service {
	return &Service{Downloaders: downloaders, Storage: storage}
}

//go:generate minimock -i Downloader -o ./mocks/downloader_mock.go -g
type Downloader interface {
	// Matches tells whether the downloader can handle the given URL.
	Matches(url string) bool
	// GetMetadata returns the metadata for the given URL.
	GetMetadata(ctx context.Context, url string) (*Metadata, error)
}

//go:generate minimock -i Storage -o ./mocks/storage_mock.go -g
type Storage interface {
	GetMetadata(ctx context.Context, url string) (*Metadata, error)
	SaveMetadata(ctx context.Context, url string, metadata *Metadata) error
}

type Service struct {
	Downloaders []Downloader
	Storage     Storage
}

func (svc *Service) getDownloader(url string) (Downloader, error) {
	for _, d := range svc.Downloaders {
		if d.Matches(url) {
			return d, nil
		}
	}
	return nil, errors.New("no downloader found for url: " + url)
}
