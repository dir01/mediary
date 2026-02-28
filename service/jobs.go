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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
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
	ctx, span := otel.Tracer("github.com/dir01/mediary/service").Start(ctx, "service.CreateJob",
		trace.WithAttributes(
			attribute.String("job.type", params.Type),
			attribute.String("job.url", params.URL),
		),
	)
	defer span.End()

	jobID := svc.calculateJobId(params)
	span.SetAttributes(attribute.String("job.id", jobID))

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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// disallow duplicate jobs
	if existingState, err := svc.storage.GetJob(ctx, jobID); err != nil {
		svc.log.Error("failed to get job state", slog.Any("error", err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to get existing job state: %w", err)
	} else if existingState != nil {
		svc.log.Debug("job already exists", slog.String("jobID", jobID))
		span.SetAttributes(attribute.Bool("job.already_exists", true))
		return existingState, errJobAlreadyExists
	}

	if err := svc.storage.SaveJob(ctx, jobState); err != nil {
		svc.log.Error("failed to save job state", slog.String("jobID", jobID), slog.Any("error", err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	svc.log.Debug("publishing job", slog.String("jobID", jobID))
	if err := svc.jobsQueue.Publish(ctx, "process", jobState.ID); err != nil {
		svc.log.Debug("failed to publish job", slog.String("jobID", jobID), slog.Any("error", err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	svc.jobsCreated.Add(ctx, 1, metric.WithAttributes(attribute.String("job.type", params.Type)))

	return jobState, nil
}

func (svc *Service) GetJob(ctx context.Context, id string) (*Job, error) {
	ctx, span := otel.Tracer("github.com/dir01/mediary/service").Start(ctx, "service.GetJob",
		trace.WithAttributes(attribute.String("job.id", id)),
	)
	defer span.End()

	job, err := svc.storage.GetJob(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	if job != nil {
		span.SetAttributes(attribute.String("job.status", job.DisplayStatus))
	}
	return job, err
}

// onPublishedJob is a callback that is invoked when a job is published to the jobs queue
// actual job is done by the corresponding flow
func (svc *Service) onPublishedJob(ctx context.Context, payload []byte) error {
	var jobID string
	if err := json.Unmarshal(payload, &jobID); err != nil {
		return fmt.Errorf("failed to unmarshal job id: %w", err)
	}
	svc.log.Debug("started onPublishedJob", slog.String("jobID", jobID))

	ctx, span := otel.Tracer("github.com/dir01/mediary/service").Start(
		ctx, "service.ProcessJob",
		trace.WithAttributes(attribute.String("job.id", jobID)),
	)
	defer span.End()

	jobState, err := svc.storage.GetJob(ctx, jobID)
	if err != nil {
		svc.log.Error("failed to get job state", slog.String("jobID", jobID), slog.Any("error", err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to get job state: %w", err)
	}
	if jobState != nil {
		span.SetAttributes(attribute.String("job.type", jobState.Type))
	}

	flow, err := svc.constructFlow(jobID, jobState)
	if err != nil {
		svc.log.Debug("failed to construct flow", slog.String("jobID", jobID), slog.Any("error", err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to construct flow: %w", err)
	}

	start := time.Now()
	if err := flow(ctx); err != nil {
		svc.log.Error("failed to execute flow", slog.String("jobID", jobID), slog.Any("error", err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		svc.jobsCompleted.Add(ctx, 1, metric.WithAttributes(
			attribute.String("job.type", jobState.Type),
			attribute.Bool("success", false),
		))
		return fmt.Errorf("failed to execute flow: %w", err)
	}

	svc.jobDuration.Record(ctx, time.Since(start).Seconds(),
		metric.WithAttributes(attribute.String("job.type", jobState.Type)),
	)
	svc.jobsCompleted.Add(ctx, 1, metric.WithAttributes(
		attribute.String("job.type", jobState.Type),
		attribute.Bool("success", true),
	))

	return nil
}

// constructFlow will return a function that will execute the given job.
// error returned if matching flow is not found of job params do not make sense
func (svc *Service) constructFlow(jobID string, jobState *Job) (func(ctx context.Context) error, error) {
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
