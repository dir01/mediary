package service

import (
	"context"
	"fmt"
	"github.com/hori-ryota/zaperr"
	"go.uber.org/zap"
	"time"
)

var ErrUrlNotSupported = fmt.Errorf("url not supported")

type Metadata struct {
	URL                   string            `json:"url"`
	Name                  string            `json:"name"`
	Variants              []VariantMetadata `json:"variants"`
	AllowMultipleVariants bool              `json:"allow_multiple_variants"`
	DownloaderName        string            `json:"downloader_name"`
}

type VariantMetadata struct {
	ID       string `json:"id"`
	LenBytes *int64 `json:"length_bytes,omitempty"`
}

func (svc *Service) GetMetadata(ctx context.Context, url string) (*Metadata, error) {
	var metadata *Metadata
	var err error
	done := make(chan struct{})

	svc.execSynced(url, func() {
		ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Minute)
		defer cancel()
		metadata, err = svc.doGetMetadata(ctx, url)
		close(done)
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return metadata, err
	}
}

func (svc *Service) doGetMetadata(ctx context.Context, url string) (*Metadata, error) {
	zapFields := []zap.Field{zap.String("url", url)}
	svc.log.Debug("getting metadata", zapFields...)

	if metadata, err := svc.storage.GetMetadata(ctx, url); err != nil {
		svc.log.Error(
			"error getting metadata from storage, will continue",
			append([]zap.Field{zaperr.ToField(err)}, zapFields...)...,
		)
	} else if metadata != nil {
		svc.log.Debug("got metadata from storage", zap.Any("metadata", metadata))
		return metadata, nil
	}

	if !svc.downloader.AcceptsURL(url) {
		return nil, zaperr.Wrap(ErrUrlNotSupported, "failed to get metadata", zapFields...)
	}

	svc.log.Debug("fetching metadata from downloader", zap.String("url", url))
	metadata, err := svc.downloader.GetMetadata(ctx, url)
	if err != nil {
		svc.log.Error("error getting metadata from downloader", append([]zap.Field{zaperr.ToField(err)}, zapFields...)...)
		return nil, zaperr.Wrap(err, "error getting metadata from downloader", zapFields...)
	}

	if err := svc.storage.SaveMetadata(ctx, metadata); err != nil {
		svc.log.Error(
			"error saving metadata to storage, will continue",
			append([]zap.Field{zaperr.ToField(err)}, zapFields...)...,
		)
	}

	return metadata, nil
}
