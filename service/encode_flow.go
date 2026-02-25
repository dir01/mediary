package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/samber/oops"
)

func (svc *Service) newUploadOriginalFlow(jobID string, job *Job) (func() error, error) {
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

	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		job, err := svc.storage.GetJob(ctx, jobID)
		if err != nil {
			return errCtx.Wrapf(err, "failed to get job")
		}

		updateJobStatus := func(status string) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			job.DisplayStatus = status
			if err = svc.storage.SaveJob(ctx, job); err != nil {
				attrs := append([]any{
					slog.String("state", job.DisplayStatus),
					slog.Any("error", err),
				}, logAttrs...)
				svc.log.Error("failed to save job state, proceeding", attrs...)
			}
		}

		updateJobStatus(JobStatusDownloading)
		svc.log.Debug("starting download", logAttrs...)
		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()
		filepathsMap, err := svc.downloader.Download(ctx, job.URL, []string{params.Variant})
		if err != nil {
			return errCtx.Wrapf(err, "failed to download files")
		}
		downloadedFilepath := filepathsMap[params.Variant]
		logAttrs = append(logAttrs, slog.String("downloadedFilepath", downloadedFilepath))
		errCtx = errCtx.With("downloadedFilepath", downloadedFilepath)
		svc.log.Debug("downloaded file", logAttrs...)

		info, err := svc.mediaProcessor.GetInfo(ctx, downloadedFilepath)
		if err != nil {
			return errCtx.Wrapf(err, "failed to get info about result file")
		}
		logAttrs = append(logAttrs, slog.Any("info", info))
		errCtx = errCtx.With("info", info)
		svc.log.Debug("got info about result file", logAttrs...)
		job.ResultMediaDuration = info.Duration
		job.ResultFileBytes = info.FileLenBytes

		updateJobStatus(JobStatusUploading)
		svc.log.Debug("starting upload", logAttrs...)
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		err = svc.uploader.Upload(ctx, downloadedFilepath, params.UploadURL)
		if err != nil {
			return errCtx.Wrapf(err, "failed to upload result")
		}

		updateJobStatus(JobStatusComplete)
		svc.log.Debug("job complete", logAttrs...)
		return nil
	}, nil
}
