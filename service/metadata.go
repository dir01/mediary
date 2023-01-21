package service

import (
	"context"
	"fmt"

	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
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
	zapFields := []zap.Field{zap.String("url", url)}
	if metadata, err := svc.storage.GetMetadata(ctx, url); err != nil {
		svc.log.Error(
			"error getting metadata from storage, will continue",
			append([]zap.Field{zap.Error(err)}, zapFields...)...,
		)
	} else if metadata != nil {
		return metadata, nil
	}

	if !svc.downloader.AcceptsURL(url) {
		return nil, zaperr.Wrap(ErrUrlNotSupported, "url not supported", zapFields...)
	}

	metadata, err := svc.downloader.GetMetadata(ctx, url)
	if err != nil {
		return nil, zaperr.Wrap(err, "error getting metadata from downloader", zapFields...)
	}

	if err := svc.storage.SaveMetadata(ctx, metadata); err != nil {
		svc.log.Error(
			"error saving metadata to storage, will continue",
			append([]zap.Field{zap.Error(err)}, zapFields...)...,
		)
	}

	return metadata, nil
}
