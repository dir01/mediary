package service

import (
	"context"
	"github.com/hori-ryota/zaperr"
	"time"

	"go.uber.org/zap"
)

func (svc *Service) newUploadOriginalFlow(jobID string, job *Job) (func() error, error) {
	zapFields := []zap.Field{
		zap.String("jobID", jobID),
		zap.Any("job", job),
	}
	type Params struct {
		Variant   string `json:"variant"`
		UploadURL string `json:"uploadUrl"`
	}
	params := Params{}
	err := mapToStruct(job.Params, &params)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to parse job params", zapFields...)
	}
	zapFields = append(zapFields, zap.Any("params", params))
	svc.log.Debug("parsed job params", zapFields...)

	if params.Variant == "" {
		return nil, zaperr.New("no filepath provided", zapFields...)
	}

	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		job, err := svc.storage.GetJob(ctx, jobID)
		if err != nil {
			return zaperr.Wrap(err, "failed to get job", zapFields...)
		}

		updateJobStatus := func(status string) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			job.DisplayStatus = status
			if err = svc.storage.SaveJob(ctx, job); err != nil {
				zapFields := append([]zap.Field{
					zap.String("state", job.DisplayStatus),
					zaperr.ToField(err),
				}, zapFields...)
				svc.log.Error("failed to save job state, proceeding", zapFields...)
			}
		}

		updateJobStatus(JobStatusDownloading)
		svc.log.Debug("starting download", zapFields...)
		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()
		filepathsMap, err := svc.downloader.Download(ctx, job.URL, []string{params.Variant})
		if err != nil {
			return zaperr.Wrap(err, "failed to download files", zapFields...)
		}
		downloadedFilepath := filepathsMap[params.Variant]
		zapFields = append(zapFields, zap.String("downloadedFilepath", downloadedFilepath))
		svc.log.Debug("downloaded file", zapFields...)

		info, err := svc.mediaProcessor.GetInfo(ctx, downloadedFilepath)
		if err != nil {
			return zaperr.Wrap(err, "failed to get info about result file: %w", zapFields...)
		}
		zapFields = append(zapFields, zap.Any("info", info))
		svc.log.Debug("got info about result file", zapFields...)
		job.ResultMediaDuration = info.Duration
		job.ResultFileBytes = info.FileLenBytes

		updateJobStatus(JobStatusUploading)
		svc.log.Debug("starting upload", zapFields...)
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		err = svc.uploader.Upload(ctx, downloadedFilepath, params.UploadURL)
		if err != nil {
			return zaperr.Wrap(err, "failed to upload result", zapFields...)
		}

		updateJobStatus(JobStatusComplete)
		svc.log.Debug("job complete", zapFields...)
		return nil
	}, nil
}
