package uploader

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dir01/mediary/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// maxUploadConcurrency caps how many uploads run at once. Several large
	// (~1GB) PUTs over a constrained uplink starve each other's sockets, so S3
	// aborts the idle connections with "400 RequestTimeout". Serializing (a
	// small N) gives each upload the full pipe. Drop to 1 if contention persists.
	maxUploadConcurrency = 2
	// maxUploadAttempts bounds retries of a single upload.
	maxUploadAttempts = 4
)

func New() (*HTTPUploader, error) {
	uploader := &HTTPUploader{
		client: &http.Client{
			// No Client.Timeout on purpose: a legitimate ~1GB upload is long and
			// is bounded by the caller's context. We instead set granular
			// transport timeouts so a *stalled* connection fails fast (and is
			// then retried) instead of hanging.
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
				TLSHandshakeTimeout:   15 * time.Second,
				ResponseHeaderTimeout: 5 * time.Minute, // starts only after the body is fully sent
				ExpectContinueTimeout: 5 * time.Second,
				IdleConnTimeout:       90 * time.Second,
				ForceAttemptHTTP2:     true,
			},
		},
		sem: make(chan struct{}, maxUploadConcurrency),
	}
	var _ service.Uploader = uploader
	return uploader, nil
}

type HTTPUploader struct {
	client *http.Client
	sem    chan struct{}
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

	// Serialize uploads: concurrent large PUTs over a constrained uplink stall
	// each other into S3 idle-socket ("RequestTimeout") errors.
	select {
	case u.sem <- struct{}{}:
		defer func() { <-u.sem }()
	case <-ctx.Done():
		return ctx.Err()
	}

	var lastErr error
	for attempt := 1; attempt <= maxUploadAttempts; attempt++ {
		if attempt > 1 {
			backoff := uploadBackoff(attempt)
			span.AddEvent("retrying upload", trace.WithAttributes(
				attribute.Int("attempt", attempt),
				attribute.String("backoff", backoff.String()),
				attribute.String("prev_error", lastErr.Error()),
			))
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				err = fmt.Errorf("upload aborted during backoff after %d attempts: %w (last error: %v)", attempt-1, ctx.Err(), lastErr)
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				return err
			}
		}

		retryable, attemptErr := u.uploadOnce(ctx, filepath, url, fileStat.Size())
		if attemptErr == nil {
			if attempt > 1 {
				span.AddEvent("upload succeeded after retry", trace.WithAttributes(attribute.Int("attempt", attempt)))
			}
			return nil
		}
		lastErr = attemptErr
		if !retryable {
			span.RecordError(attemptErr)
			span.SetStatus(codes.Error, attemptErr.Error())
			return attemptErr
		}
	}

	err = fmt.Errorf("upload failed after %d attempts: %w", maxUploadAttempts, lastErr)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}

// uploadOnce performs a single PUT attempt and reports whether a failure is
// worth retrying.
func (u *HTTPUploader) uploadOnce(ctx context.Context, filepath, url string, size int64) (retryable bool, err error) {
	file, err := os.Open(filepath)
	if err != nil {
		return false, fmt.Errorf("error opening file: %w", err)
	}
	defer func() { _ = file.Close() }()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, file)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.ContentLength = size

	resp, err := u.client.Do(req)
	if err != nil {
		// Transport-level failures (stalled/closed connections, dial failures)
		// are transient; retry unless the context itself is done.
		if ctx.Err() != nil {
			return false, fmt.Errorf("failed sending request: %w", err)
		}
		return true, fmt.Errorf("failed sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return false, nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
	err = fmt.Errorf("unexpected status code: %d (response body: %s)", resp.StatusCode, string(body))
	return isRetryableStatus(resp.StatusCode, body), err
}

// isRetryableStatus reports whether an HTTP response warrants a retry. S3's
// "RequestTimeout" is a 400 but is explicitly transient (it closed an idle
// socket), so it is retried; other 4xx (e.g. 403 on an expired URL) are not.
func isRetryableStatus(status int, body []byte) bool {
	switch status {
	case http.StatusRequestTimeout, // 408
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	case http.StatusBadRequest: // 400
		return strings.Contains(string(body), "RequestTimeout")
	default:
		return false
	}
}

// uploadBackoff returns an exponential backoff with jitter for the given
// attempt (attempt 2 -> ~2s, 3 -> ~4s, 4 -> ~8s, capped at 30s).
func uploadBackoff(attempt int) time.Duration {
	const base = 2 * time.Second
	const maxDelay = 30 * time.Second
	d := base << (attempt - 2)
	if d > maxDelay {
		d = maxDelay
	}
	return d + time.Duration(rand.Int64N(int64(d/2))) // up to 50% jitter
}
