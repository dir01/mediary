package uploader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dir01/mediary/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func New() (*HTTPUploader, error) {
	uploader := &HTTPUploader{}
	var _ service.Uploader = uploader
	return uploader, nil
}

type HTTPUploader struct {
}

func (u *HTTPUploader) Upload(ctx context.Context, filepath string, url string) error {
	ctx, span := otel.Tracer("github.com/dir01/mediary/uploader").Start(ctx, "uploader.Upload",
		trace.WithAttributes(attribute.String("upload.url", url)),
	)
	defer span.End()

	fileStat, err := os.Stat(filepath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to get file stat: %w", err)
	}
	span.SetAttributes(attribute.Int64("upload.bytes", fileStat.Size()))

	file, err := os.Open(filepath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("error opening file: %w", err)
	}
	defer func() { _ = file.Close() }()

	req, err := http.NewRequestWithContext(ctx, "PUT", url, file)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.ContentLength = fileStat.Size()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	span.SetAttributes(attribute.Int("http.response.status_code", resp.StatusCode))

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("unexpected status code: %d (failed to read response body: %w)", resp.StatusCode, err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	err = fmt.Errorf("unexpected status code: %d (response body: %s)", resp.StatusCode, string(bytes))
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}
