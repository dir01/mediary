package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

var (
	errUnsupportedJobType = fmt.Errorf("unsupported job type")
	errJobAlreadyExists   = fmt.Errorf("job already exists")
	errInvalidJobParams   = fmt.Errorf("invalid job params")
)

type JobParams struct {
	URL    string                 `json:"url"`
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

const (
	jobTypeConcatenate = "concatenate"
)

type JobState struct {
	JobParams
	ID            string `json:"id"`
	DisplayStatus string `json:"status"`
}

const StatusCreated = "created"
const StatusDownloading = "downloading"
const StatusProcessing = "processing"
const StatusUploading = "uploading"
const StatusComplete = "complete"

// CreateJob creates an entry for job in storage and enqueues it for processing in background
func (svc *Service) CreateJob(ctx context.Context, params *JobParams) (*JobState, error) {
	jobID := svc.calculateJobId(params)
	svc.log.Debug(
		"started CreateJob",
		zap.String("jobID", jobID),
		zap.Any("params", params),
	)
	jobState := &JobState{
		JobParams:     *params,
		ID:            jobID,
		DisplayStatus: StatusCreated,
	}

	// rough validation of job params
	if _, err := svc.constructFlow(jobID, jobState); err != nil {
		return nil, err
	}

	// disallow duplicate jobs
	if existingstate, err := svc.storage.GetJob(ctx, jobID); err != nil {
		svc.log.Error("failed to get job state", zap.String("jobID", jobID), zap.Error(err))
		return nil, fmt.Errorf("failed to get existing job state: %w", err)
	} else if existingstate != nil {
		svc.log.Debug("job already exists", zap.String("jobID", jobID))
		return existingstate, errJobAlreadyExists
	}

	if err := svc.storage.SaveJob(ctx, jobState); err != nil {
		svc.log.Error("failed to save job state", zap.String("jobID", jobID), zap.Error(err))
		return nil, err
	}

	svc.log.Debug("publishing job", zap.String("jobID", jobID))
	if err := svc.jobsQueue.Publish(ctx, jobState.ID); err != nil {
		svc.log.Debug("failed to publish job", zap.String("jobID", jobID), zap.Error(err))
		return nil, err
	}

	return jobState, nil
}

func (svc *Service) GetJob(ctx context.Context, id string) (*JobState, error) {
	return svc.storage.GetJob(ctx, id)
}

// onPublishedJob is a callback that is invoked when a job is published to the jobs queue
// actual job is done by the corresponding flow
func (svc *Service) onPublishedJob(jobID string) error {
	svc.log.Debug("started onPublishedJob", zap.String("jobID", jobID))

	jobState, err := svc.storage.GetJob(context.Background(), jobID)
	if err != nil {
		svc.log.Error("failed to get job state", zap.String("jobID", jobID), zap.Error(err))
		return fmt.Errorf("failed to get job state: %w", err)
	}

	flow, err := svc.constructFlow(jobID, jobState)
	if err != nil {
		svc.log.Debug("failed to construct flow", zap.String("jobID", jobID), zap.Error(err))
		return fmt.Errorf("failed to construct flow: %w", err)
	}

	if err := flow(); err != nil {
		svc.log.Error("failed to execute flow", zap.String("jobID", jobID), zap.Error(err))
		return fmt.Errorf("failed to execute flow: %w", err)
	}

	return nil
}

// constructFlow will return a function that will execute the given job.
// error returned if matching flow is not found of job params do not make sense
func (svc *Service) constructFlow(jobID string, jobState *JobState) (func() error, error) {
	switch jobState.Type {
	case jobTypeConcatenate:
		return svc.newConcatenateFlow(jobID, jobState)
	default:
		return nil, errUnsupportedJobType
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
