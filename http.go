package mediary

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

func PrepareHTTPServerMux(service *Service) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/metadata", handleMetadata(service, 100*time.Millisecond))
	mux.HandleFunc("/metadata/long-polling", handleMetadata(service, 5*time.Minute))
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
		} else if errors.Is(err, context.DeadlineExceeded) {
			respondWithJSON(w, http.StatusAccepted, `{"foo": "bar"}`)
		} else {
			respondWithJSON(w, http.StatusInternalServerError, err.Error())
		}
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	var response []byte
	switch payload := payload.(type) {
	case error:
		response = []byte(payload.Error())
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
