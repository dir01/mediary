package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/dir01/mediary/service"
)

// extractURLParam extracts the "url" query parameter from a GET request.
// Magnet URLs contain literal '&' separating parameters (e.g. &tr=, &dn=)
// which standard query parsing splits on, truncating the URL.
// When the parsed value looks like a magnet, we extract the raw value instead.
func extractURLParam(req *http.Request) string {
	url := req.URL.Query().Get("url")
	if strings.HasPrefix(url, "magnet:") {
		if i := strings.Index(req.URL.RawQuery, "url="); i != -1 {
			if raw, err := neturl.QueryUnescape(req.URL.RawQuery[i+4:]); err == nil {
				return raw
			}
		}
	}
	return url
}

func handleGetMetadata(svc *service.Service, timeout time.Duration) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if timeout != 0 {
			ctx1, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			ctx = ctx1
		}

		var url string
		switch req.Method {
		case http.MethodGet:
			url = extractURLParam(req)
		case http.MethodPost:
			// read json body
			var body struct {
				URL string `json:"url"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				respond(w, http.StatusBadRequest, err)
				return
			}
			url = body.URL
		default:
			respond(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		}
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
