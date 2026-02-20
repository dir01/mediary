package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

func (svc *Service) newUploadOriginalFlow(jobID string, job *Job) (func() error, error) {
	logAttrs := []any{
		slog.String("jobID", jobID),
		slog.Any("job", job),
	}
	type Params struct {
		Variant   string `json:"variant"`
		UploadURL string `json:"uploadUrl"`
	}
	params := Params{}
	err := mapToStruct(job.Params, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse job params: %w", err)
	}
	logAttrs = append(logAttrs, slog.Any("params", params))
	svc.log.Debug("parsed job params", logAttrs...)

	if params.Variant == "" {
		return nil, fmt.Errorf("no filepath provided")
	}

	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		job, err := svc.storage.GetJob(ctx, jobID)
		if err != nil {
			return fmt.Errorf("failed to get job: %w", err)
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
			return fmt.Errorf("failed to download files: %w", err)
		}
		downloadedFilepath := filepathsMap[params.Variant]
		logAttrs = append(logAttrs, slog.String("downloadedFilepath", downloadedFilepath))
		svc.log.Debug("downloaded file", logAttrs...)

		info, err := svc.mediaProcessor.GetInfo(ctx, downloadedFilepath)
		if err != nil {
			return fmt.Errorf("failed to get info about result file: %w", err)
		}
		logAttrs = append(logAttrs, slog.Any("info", info))
		svc.log.Debug("got info about result file", logAttrs...)
		job.ResultMediaDuration = info.Duration
		job.ResultFileBytes = info.FileLenBytes

		updateJobStatus(JobStatusUploading)
		svc.log.Debug("starting upload", logAttrs...)
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		err = svc.uploader.Upload(ctx, downloadedFilepath, params.UploadURL)
		if err != nil {
			return fmt.Errorf("failed to upload result: %w", err)
		}

		updateJobStatus(JobStatusComplete)
		svc.log.Debug("job complete", logAttrs...)
		return nil
	}, nil
}
