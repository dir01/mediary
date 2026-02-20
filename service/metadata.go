package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/oops"
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
	logAttrs := []any{slog.String("url", url)}
	errCtx := oops.With("url", url)
	svc.log.Debug("getting metadata", logAttrs...)

	if metadata, err := svc.storage.GetMetadata(ctx, url); err != nil {
		svc.log.Error(
			"error getting metadata from storage, will continue",
			append([]any{slog.Any("error", err)}, logAttrs...)...,
		)
	} else if metadata != nil {
		svc.log.Debug("got metadata from storage", slog.Any("metadata", metadata))
		return metadata, nil
	}

	if !svc.downloader.AcceptsURL(url) {
		return nil, errCtx.Wrapf(ErrUrlNotSupported, "failed to get metadata")
	}

	svc.log.Debug("fetching metadata from downloader", slog.String("url", url))
	metadata, err := svc.downloader.GetMetadata(ctx, url)
	if err != nil {
		svc.log.Error("error getting metadata from downloader", append([]any{slog.Any("error", err)}, logAttrs...)...)
		return nil, errCtx.Wrapf(err, "error getting metadata from downloader")
	}

	if err := svc.storage.SaveMetadata(ctx, metadata); err != nil {
		svc.log.Error(
			"error saving metadata to storage, will continue",
			append([]any{slog.Any("error", err)}, logAttrs...)...,
		)
	}

	return metadata, nil
}
