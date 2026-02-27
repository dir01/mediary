package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/samber/oops"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (svc *Service) newUploadOriginalFlow(jobID string, job *Job) (func(ctx context.Context) error, error) {
	logAttrs := []any{
		slog.String("jobID", jobID),
		slog.Any("job", job),
	}
	errCtx := oops.With("jobID", jobID, "job", job)
	type Params struct {
		Variant   string `json:"variant"`
		UploadURL string `json:"uploadUrl"`
	}
	params := Params{}
	err := mapToStruct(job.Params, &params)
	if err != nil {
		return nil, errCtx.Wrapf(err, "failed to parse job params")
	}
	logAttrs = append(logAttrs, slog.Any("params", params))
	errCtx = errCtx.With("params", params)
	svc.log.Debug("parsed job params", logAttrs...)

	if params.Variant == "" {
		return nil, errCtx.Errorf("no filepath provided")
	}

	return func(jobCtx context.Context) error {
		jobCtx, span := otel.Tracer("github.com/dir01/mediary/service").Start(jobCtx, "service.UploadOriginalFlow",
			trace.WithAttributes(
				attribute.String("job.id", jobID),
				attribute.String("variant", params.Variant),
			),
		)
		defer span.End()

		ctx, cancel := context.WithTimeout(jobCtx, 1*time.Second)
		defer cancel()
		job, err := svc.storage.GetJob(ctx, jobID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return errCtx.Wrapf(err, "failed to get job")
		}

		updateJobStatus := func(status string) {
			statusCtx, statusCancel := context.WithTimeout(jobCtx, 1*time.Second)
			defer statusCancel()
			job.DisplayStatus = status
			if err = svc.storage.SaveJob(statusCtx, job); err != nil {
				attrs := append([]any{
					slog.String("state", job.DisplayStatus),
					slog.Any("error", err),
				}, logAttrs...)
				svc.log.Error("failed to save job state, proceeding", attrs...)
			}
		}

		updateJobStatus(JobStatusDownloading)
		svc.log.Debug("starting download", logAttrs...)

		downloadCtx, downloadCancel := context.WithTimeout(jobCtx, 1*time.Hour)
		defer downloadCancel()

		downloadCtx, downloadSpan := otel.Tracer("github.com/dir01/mediary/service").Start(downloadCtx, "service.Download",
			trace.WithAttributes(
				attribute.String("job.id", jobID),
				attribute.String("url", job.URL),
				attribute.String("variant", params.Variant),
			),
		)
		filepathsMap, err := svc.downloader.Download(downloadCtx, job.URL, []string{params.Variant})
		if err != nil {
			downloadSpan.RecordError(err)
			downloadSpan.SetStatus(codes.Error, err.Error())
			downloadSpan.End()
			return errCtx.Wrapf(err, "failed to download files")
		}
		downloadSpan.End()

		downloadedFilepath := filepathsMap[params.Variant]
		logAttrs = append(logAttrs, slog.String("downloadedFilepath", downloadedFilepath))
		errCtx = errCtx.With("downloadedFilepath", downloadedFilepath)
		svc.log.Debug("downloaded file", logAttrs...)

		info, err := svc.mediaProcessor.GetInfo(downloadCtx, downloadedFilepath)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return errCtx.Wrapf(err, "failed to get info about result file")
		}
		logAttrs = append(logAttrs, slog.Any("info", info))
		errCtx = errCtx.With("info", info)
		svc.log.Debug("got info about result file", logAttrs...)
		job.ResultMediaDuration = info.Duration
		job.ResultFileBytes = info.FileLenBytes
		span.SetAttributes(
			attribute.Int64("result.bytes", info.FileLenBytes),
			attribute.Float64("result.duration_seconds", info.Duration.Seconds()),
		)

		updateJobStatus(JobStatusUploading)
		svc.log.Debug("starting upload", logAttrs...)

		uploadCtx, uploadCancel := context.WithTimeout(jobCtx, 30*time.Minute)
		defer uploadCancel()

		err = svc.uploader.Upload(uploadCtx, downloadedFilepath, params.UploadURL)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return errCtx.Wrapf(err, "failed to upload result")
		}

		updateJobStatus(JobStatusComplete)
		svc.log.Debug("job complete", logAttrs...)
		return nil
	}, nil
}
