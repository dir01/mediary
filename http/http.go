package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dir01/mediary/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func PrepareHTTPServerMux(service *service.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/metadata", handleGetMetadata(service, 100*time.Millisecond))
	mux.HandleFunc("/metadata/long-polling", handleGetMetadata(service, 5*time.Minute))
	mux.HandleFunc("/jobs/", handleGetJob(service))
	mux.HandleFunc("/jobs", handleCreateJob(service))
	mux.HandleFunc("/", handleDocs())
	return otelhttp.NewHandler(mux, "mediary",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			path := r.URL.Path
			if strings.HasPrefix(path, "/jobs/") && len(path) > len("/jobs/") {
				path = "/jobs/:id"
			}
			return r.Method + " " + path
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
