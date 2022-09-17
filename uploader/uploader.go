package uploader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dir01/mediary/service"
)

func New() (*HTTPUploader, error) {
	uploader := &HTTPUploader{}
	var _ service.Uploader = uploader
	return uploader, nil
}

type HTTPUploader struct {
}

func (u *HTTPUploader) Upload(ctx context.Context, filepath string, url string) error {
	fileStat, err := os.Stat(filepath)
	if err != nil {
		return fmt.Errorf("failed to get file stat: %w", err)
	}

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer func() { _ = file.Close() }()

	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.ContentLength = fileStat.Size()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unexpected status code: %d (failed to read response body: %w)", resp.StatusCode, err)
	}

	return fmt.Errorf("unexpected status code: %d (response body: %s)", resp.StatusCode, string(bytes))
}
