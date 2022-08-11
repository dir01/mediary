package mediary

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func PrepareHTTPServerMux(service *Service) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/metadata", handleMetadata(service, 5*time.Second))
	mux.HandleFunc("/metadata/long-polling", handleMetadata(service, 0))
	return mux
}

func handleMetadata(svc *Service, timeout time.Duration) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if timeout != 0 {
			ctx1, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			ctx = ctx1
		}

		if metadata, err := svc.GetMetadata(ctx, req.URL.Query().Get("url")); err == nil {
			respondWithJSON(w, http.StatusOK, metadata)
		} else {
			respondWithJSON(w, http.StatusInternalServerError, err.Error())
		}
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}
