package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dir01/mediary/service"
)

func PrepareHTTPServerMux(service *service.Service) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/metadata", handleGetMetadata(service, 100*time.Millisecond))
	mux.HandleFunc("/metadata/long-polling", handleGetMetadata(service, 5*time.Minute))
	mux.HandleFunc("/jobs/", handleGetJob(service))
	mux.HandleFunc("/jobs", handleCreateJob(service))
	//mux.HandleFunc("/", handleDocs())
	return mux
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
	return
}
