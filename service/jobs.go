package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/samber/oops"
)

var (
	errUnsupportedJobType = fmt.Errorf("unsupported job type")
	errJobAlreadyExists   = fmt.Errorf("job already exists")
)

type JobParams struct {
	ID     string                 `json:"id"`
	URL    string                 `json:"url"`
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

const (
	jobTypeConcatenate    = "concatenate"
	jobTypeUploadOriginal = "upload_original"
)

type Job struct {
	JobParams
	ID                  string        `json:"id"`
	DisplayStatus       string        `json:"status"`
	ResultMediaDuration time.Duration `json:"result_media_duration,omitempty"`
	ResultFileBytes     int64         `json:"result_file_bytes,omitempty"`
}

const JobStatusCreated = "created"
const JobStatusDownloading = "downloading"
const JobStatusProcessing = "processing"
const JobStatusUploading = "uploading"
const JobStatusComplete = "complete"

// CreateJob creates an entry for job in storage and enqueues it for processing in background
func (svc *Service) CreateJob(ctx context.Context, params *JobParams) (*Job, error) {
	jobID := svc.calculateJobId(params)

	logAttrs := []any{
		slog.String("jobID", jobID),
		slog.Any("params", params),
	}
	svc.log.Debug("started CreateJob", logAttrs...)

	jobState := &Job{
		JobParams:     *params,
		ID:            jobID,
		DisplayStatus: JobStatusCreated,
	}

	// rough validation of job params
	if _, err := svc.constructFlow(jobID, jobState); err != nil {
		return nil, err
	}

	// disallow duplicate jobs
	if existingState, err := svc.storage.GetJob(ctx, jobID); err != nil {
		svc.log.Error("failed to get job state", slog.Any("error", err))
		return nil, fmt.Errorf("failed to get existing job state: %w", err)
	} else if existingState != nil {
		svc.log.Debug("job already exists", slog.String("jobID", jobID))
		return existingState, errJobAlreadyExists
	}

	if err := svc.storage.SaveJob(ctx, jobState); err != nil {
		svc.log.Error("failed to save job state", slog.String("jobID", jobID), slog.Any("error", err))
		return nil, err
	}

	svc.log.Debug("publishing job", slog.String("jobID", jobID))
	if err := svc.jobsQueue.Publish(ctx, "process", jobState.ID); err != nil {
		svc.log.Debug("failed to publish job", slog.String("jobID", jobID), slog.Any("error", err))
		return nil, err
	}

	return jobState, nil
}

func (svc *Service) GetJob(ctx context.Context, id string) (*Job, error) {
	return svc.storage.GetJob(ctx, id)
}

// onPublishedJob is a callback that is invoked when a job is published to the jobs queue
// actual job is done by the corresponding flow
func (svc *Service) onPublishedJob(payload []byte) error {
	var jobID string
	if err := json.Unmarshal(payload, &jobID); err != nil {
		return fmt.Errorf("failed to unmarshal job id: %w", err)
	}
	svc.log.Debug("started onPublishedJob", slog.String("jobID", jobID))

	jobState, err := svc.storage.GetJob(context.Background(), jobID)
	if err != nil {
		svc.log.Error("failed to get job state", slog.String("jobID", jobID), slog.Any("error", err))
		return fmt.Errorf("failed to get job state: %w", err)
	}

	flow, err := svc.constructFlow(jobID, jobState)
	if err != nil {
		svc.log.Debug("failed to construct flow", slog.String("jobID", jobID), slog.Any("error", err))
		return fmt.Errorf("failed to construct flow: %w", err)
	}

	if err := flow(); err != nil {
		svc.log.Error("failed to execute flow", slog.String("jobID", jobID), slog.Any("error", err))
		return fmt.Errorf("failed to execute flow: %w", err)
	}

	return nil
}

// constructFlow will return a function that will execute the given job.
// error returned if matching flow is not found of job params do not make sense
func (svc *Service) constructFlow(jobID string, jobState *Job) (func() error, error) {
	switch jobState.Type {
	case jobTypeConcatenate:
		return svc.newConcatenateFlow(jobID, jobState)
	case jobTypeUploadOriginal:
		return svc.newUploadOriginalFlow(jobID, jobState)
	default:
		return nil, oops.With("jobType", jobState.Type).Wrapf(errUnsupportedJobType, "unsupported job type: %s", jobState.Type)
	}
}

// calculateJobId returns a unique identifier for the given job parameters.
// for a given set of parameters, it will always return the same job id
func (svc *Service) calculateJobId(params *JobParams) string {
	// TODO: use stable json serialization, currently there are no guarantees that same params will result in same id
	bytes, _ := json.Marshal(params)
	hash := md5.Sum(bytes)
	return hex.EncodeToString(hash[:])
}
