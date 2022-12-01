package service

import (
	"context"
	"fmt"
	"log"
)

var ErrUrlNotSupported = fmt.Errorf("url not supported")

type Metadata struct {
	URL   string         `json:"url"`
	Name  string         `json:"name"`
	Files []FileMetadata `json:"files"`
}

type FileMetadata struct {
	Path     string `json:"path"`
	LenBytes int64  `json:"length_bytes"`
}

func (svc *Service) GetMetadata(ctx context.Context, url string) (*Metadata, error) {
	if metadata, err := svc.storage.GetMetadata(ctx, url); err != nil {
		log.Printf("error getting metadata from storage: %v, will continue", err)
	} else if metadata != nil {
		return metadata, nil
	}

	if !svc.downloader.AcceptsURL(url) {
		return nil, ErrUrlNotSupported
	}

	metadata, err := svc.downloader.GetMetadata(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("error getting metadata from downloader: %w", err)
	}

	if err := svc.storage.SaveMetadata(ctx, metadata); err != nil {
		fmt.Printf("error saving metadata to storage: %v, will continue", err)
	}

	return metadata, nil
}
