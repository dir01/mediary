package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/oops"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	tracer := otel.Tracer("github.com/dir01/mediary/service")
	ctx, span := tracer.Start(ctx, "service.GetMetadata",
		trace.WithAttributes(attribute.String("url", url)),
	)
	defer span.End()

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
		span.SetAttributes(attribute.Bool("metadata.cached", true))
		return metadata, nil
	}

	if !svc.downloader.AcceptsURL(url) {
		err := errCtx.Wrapf(ErrUrlNotSupported, "failed to get metadata")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	svc.log.Debug("fetching metadata from downloader", slog.String("url", url))
	metadata, err := svc.downloader.GetMetadata(ctx, url)
	if err != nil {
		svc.log.Error("error getting metadata from downloader", append([]any{slog.Any("error", err)}, logAttrs...)...)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errCtx.Wrapf(err, "error getting metadata from downloader")
	}

	span.SetAttributes(
		attribute.String("metadata.downloader", metadata.DownloaderName),
		attribute.Int("metadata.variant_count", len(metadata.Variants)),
	)

	if err := svc.storage.SaveMetadata(ctx, metadata); err != nil {
		svc.log.Error(
			"error saving metadata to storage, will continue",
			append([]any{slog.Any("error", err)}, logAttrs...)...,
		)
	}

	return metadata, nil
}
