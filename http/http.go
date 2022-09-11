package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dir01/mediary/service"
)

func PrepareHTTPServerMux(service *service.Service) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/metadata", handleGetMetadata(service, 100*time.Millisecond))
	mux.HandleFunc("/metadata/long-polling", handleGetMetadata(service, 5*time.Minute))
	mux.HandleFunc("/jobs", handleCreateJob(service))
	mux.HandleFunc("/", handleDocs())
	return mux
}

func handleCreateJob(svc *service.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		params := &service.JobParams{}
		if err := json.NewDecoder(req.Body).Decode(params); err != nil {
			respond(w, http.StatusBadRequest, fmt.Errorf("failed to unmarshal json request body: %w", err))
			return
		}
		if _, err := svc.CreateJob(req.Context(), params); err == nil {
			respond(w, http.StatusAccepted, `{"status": "accepted"}`)
		} else {
			respond(w, http.StatusInternalServerError, fmt.Errorf("failed to create job: %w", err))
		}
	}
}

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
		} else if errors.Is(err, context.DeadlineExceeded) {
			respond(w, http.StatusAccepted, `{"status": "accepted"}`)
		} else {
			respond(w, http.StatusInternalServerError, err)
		}
	}
}

func respond(w http.ResponseWriter, code int, payload interface{}) {
	var response []byte
	switch payload := payload.(type) {
	case error:
		response = []byte(fmt.Sprintf(`{"status": "error", "error": "%s"}`, payload.Error()))
	case string:
		response = []byte(payload)
	default:
		var err error
		response, err = json.MarshalIndent(payload, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
	w.WriteHeader(code)
	w.Write(response)
}
