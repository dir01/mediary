package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
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
	// ErrJobAlreadyExists is returned by CreateJob when a job with the same
	// identity already exists; the existing job is returned alongside it.
	ErrJobAlreadyExists = fmt.Errorf("job already exists")
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

// JobStatusCancelled is a terminal status set via CancelJob. It is not produced
// by the processing flows themselves.
const JobStatusCancelled = "cancelled"

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
		return existingState, ErrJobAlreadyExists
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
	if jobState == nil {
		// The job was deleted (purged) after being enqueued but before we picked
		// it up. There is nothing to run — treat it as a successful no-op so the
		// queue does not retry (and so constructFlow is never handed a nil job,
		// which would panic). This is the expected path when a job is purged to
		// free its id for a retry.
		svc.log.Info("job no longer exists, skipping", slog.String("jobID", jobID))
		span.SetAttributes(attribute.Bool("job.missing", true))
		return nil
	}
	span.SetAttributes(attribute.String("job.type", jobState.Type))

	flow, err := svc.constructFlow(jobID, jobState)
	if err != nil {
		svc.log.Debug("failed to construct flow", slog.String("jobID", jobID), slog.Any("error", err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to construct flow: %w", err)
	}

	// Run the flow under a cancelable context registered by job id, so CancelJob
	// and DeleteJob can stop it mid-flight.
	flowCtx, cancelFlow := context.WithCancel(ctx)
	defer svc.beginJob(jobID, cancelFlow)()

	start := time.Now()
	if err := flow(flowCtx); err != nil {
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

// calculateJobId returns a stable identifier for the given job parameters.
// For a given set of parameters it always returns the same id, which is what
// makes job creation idempotent (see the duplicate check in CreateJob).
//
// The destination "uploadUrl" is normalized to just its path before hashing:
// it is a presigned S3 URL whose signature and expiry change on every request,
// so hashing it verbatim would make identical work hash to a different id each
// time — defeating de-duplication and causing the same download/convert/upload
// to run again and again (expensive). The path still encodes the destination
// (bucket + user + key), so genuinely distinct destinations stay distinct.
func (svc *Service) calculateJobId(params *JobParams) string {
	normalized := *params
	if len(params.Params) > 0 {
		p := make(map[string]interface{}, len(params.Params))
		for k, v := range params.Params {
			if k == "uploadUrl" {
				if s, ok := v.(string); ok {
					v = stripURLQuery(s)
				}
			}
			p[k] = v
		}
		normalized.Params = p
	}
	// encoding/json sorts map keys, so this is stable for a given input.
	bytes, _ := json.Marshal(&normalized)
	hash := md5.Sum(bytes)
	return hex.EncodeToString(hash[:])
}

// stripURLQuery returns the URL without its query string or fragment.
func stripURLQuery(raw string) string {
	if u, err := url.Parse(raw); err == nil {
		u.RawQuery = ""
		u.Fragment = ""
		return u.String()
	}
	if i := strings.IndexByte(raw, '?'); i >= 0 {
		return raw[:i]
	}
	return raw
}

// CancelJob marks a job terminally cancelled and stops its flow if it is
// currently running (its context is cancelled, so the active stage aborts and
// no further status writes land). It also gives a stuck or never-completing job
// a terminal state to observe. Returns (nil, nil) when the job does not exist.
func (svc *Service) CancelJob(ctx context.Context, id string) (*Job, error) {
	ctx, span := otel.Tracer("github.com/dir01/mediary/service").Start(ctx, "service.CancelJob",
		trace.WithAttributes(attribute.String("job.id", id)),
	)
	defer span.End()

	job, err := svc.storage.GetJob(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	if job == nil {
		return nil, nil
	}

	job.DisplayStatus = JobStatusCancelled
	if err := svc.storage.SaveJob(ctx, job); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to save cancelled job: %w", err)
	}
	svc.stopRunningJob(id)
	return job, nil
}

// DeleteJob stops a job's flow if it is running, then hard-removes the record so
// its content-addressed id is free again: because CreateJob dedupes on row
// existence (not status), a subsequent create with the same params re-enqueues
// real work instead of attaching to the old record. This is what makes a "retry"
// possible for a failed/cancelled job.
//
// Stopping the flow first (cancelling its context) is what keeps the delete
// durable: SaveJob is an upsert, so an un-cancelled flow's next status write
// would resurrect the row; a cancelled flow's writes fail instead.
// Returns the deleted job, or (nil, nil) when it does not exist.
func (svc *Service) DeleteJob(ctx context.Context, id string) (*Job, error) {
	ctx, span := otel.Tracer("github.com/dir01/mediary/service").Start(ctx, "service.DeleteJob",
		trace.WithAttributes(attribute.String("job.id", id)),
	)
	defer span.End()

	job, err := svc.storage.GetJob(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	if job == nil {
		return nil, nil
	}

	// Stop any in-flight flow first so its status writes fail on the cancelled
	// context instead of resurrecting the row we are about to delete.
	svc.stopRunningJob(id)

	if err := svc.storage.DeleteJob(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to delete job: %w", err)
	}
	return job, nil
}
