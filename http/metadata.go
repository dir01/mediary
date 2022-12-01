package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dir01/mediary/service"
)

func handleGetMetadata(svc *service.Service, timeout time.Duration) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if timeout != 0 {
			ctx1, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			ctx = ctx1
		}

		url := req.URL.Query().Get("url")
		if url == "" {
			respond(w, http.StatusBadRequest, errors.New("missing url parameter"))
			return
		}
		if metadata, err := svc.GetMetadata(ctx, url); err == nil {
			respond(w, http.StatusOK, metadata)
			return
		} else if errors.Is(err, context.DeadlineExceeded) {
			respond(w, http.StatusAccepted, `{"status": "accepted"}`)
			return
		} else if errors.Is(err, service.ErrUrlNotSupported) {
			respond(w, http.StatusBadRequest, fmt.Errorf("url not supported: %w", err))
			return
		} else {
			respond(w, http.StatusInternalServerError, err)
			return
		}
	}
}
