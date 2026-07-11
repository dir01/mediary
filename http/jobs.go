package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dir01/mediary/service"
)

func handleCreateJob(svc *service.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
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

// handleDeleteJob handles DELETE /jobs/{id}. By default it cancels the job
// (marks it terminally cancelled but keeps the record). With ?purge=true it
// hard-removes the record, freeing the content-addressed id so an identical
// create re-runs the work instead of attaching to the old (dead) job — the
// basis for retrying a failed job.
func handleDeleteJob(svc *service.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		id := req.PathValue("id")

		purge := req.URL.Query().Get("purge") == "true"
		action := svc.CancelJob
		verb := "cancel"
		if purge {
			action = svc.DeleteJob
			verb = "delete"
		}

		job, err := action(req.Context(), id)
		if err != nil {
			respond(w, http.StatusInternalServerError, fmt.Errorf("failed to %s job: %w", verb, err))
			return
		}
		if job == nil {
			respond(w, http.StatusNotFound, fmt.Errorf("job not found"))
			return
		}
		respond(w, http.StatusOK, job)
	}
}

func handleGetJob(svc *service.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		id := req.PathValue("id")
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
