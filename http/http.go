package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dir01/mediary/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func PrepareHTTPServerMux(service *service.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /metadata", handleGetMetadata(service, 100*time.Millisecond))
	mux.HandleFunc("POST /metadata", handleGetMetadata(service, 100*time.Millisecond))
	mux.HandleFunc("GET /metadata/long-polling", handleGetMetadata(service, 5*time.Minute))
	mux.HandleFunc("POST /metadata/long-polling", handleGetMetadata(service, 5*time.Minute))
	mux.HandleFunc("POST /jobs", handleCreateJob(service))
	mux.HandleFunc("GET /jobs/{id}", handleGetJob(service))
	mux.HandleFunc("DELETE /jobs/{id}", handleDeleteJob(service))
	mux.HandleFunc("GET /", handleDocs())
	return otelhttp.NewHandler(mux, "mediary",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			// r.Pattern is the matched route template (e.g. "GET /jobs/{id}"),
			// which is already low-cardinality — exactly what a span name should be.
			if r.Pattern != "" {
				return r.Pattern
			}
			// Fallback for requests that matched no pattern (e.g. 404s).
			return r.Method + " " + r.URL.Path
		}),
	)
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
			if _, err := w.Write([]byte(err.Error())); err != nil {
				fmt.Printf("error writing response: %s", err)
			}
			return
		}
	}
	w.WriteHeader(code)
	if _, err := w.Write(response); err != nil {
		fmt.Printf("error writing response: %s", err)
	}
}
