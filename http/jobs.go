package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dir01/mediary/service"
)

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
		job, err := svc.CreateJob(req.Context(), params)
		// An already-existing job is success: CreateJob is idempotent and returns
		// the existing job, so the caller can attach to it instead of retrying and
		// spawning duplicate (expensive) work.
		if err == nil || errors.Is(err, service.ErrJobAlreadyExists) {
			respond(w, http.StatusAccepted, fmt.Sprintf(`{"status": "accepted", "id": "%s"}`, job.ID))
			return
		}
		respond(w, http.StatusInternalServerError, fmt.Errorf("failed to create job: %w", err))
	}
}

func handleDeleteJob(svc *service.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		id := strings.TrimPrefix(req.URL.Path, "/jobs/")
		if id == "" {
			respond(w, http.StatusBadRequest, fmt.Errorf("missing job id"))
			return
		}
		job, err := svc.CancelJob(req.Context(), id)
		if err != nil {
			respond(w, http.StatusInternalServerError, fmt.Errorf("failed to cancel job: %w", err))
			return
		}
		if job == nil {
			respond(w, http.StatusNotFound, fmt.Errorf("job not found"))
			return
		}
		respond(w, http.StatusOK, job)
	}
}

// handleJobByID dispatches /jobs/{id} by method: DELETE cancels the job, any
// other method reads it.
func handleJobByID(svc *service.Service) func(http.ResponseWriter, *http.Request) {
	get := handleGetJob(svc)
	del := handleDeleteJob(svc)
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodDelete {
			del(w, req)
			return
		}
		get(w, req)
	}
}

func handleGetJob(svc *service.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		id := strings.TrimPrefix(req.URL.Path, "/jobs/")
		if id == "" {
			respond(w, http.StatusBadRequest, fmt.Errorf("missing job id"))
			return
		}
		job, err := svc.GetJob(req.Context(), id)
		if err != nil {
			respond(w, http.StatusInternalServerError, fmt.Errorf("failed to get job: %w", err))
			return
		}
		if job == nil {
			respond(w, http.StatusNotFound, fmt.Errorf("job not found"))
			return
		}
		respond(w, http.StatusOK, job)
	}
}
