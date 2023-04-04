package service

import (
	"context"
	"github.com/hori-ryota/zaperr"
	"time"

	"go.uber.org/zap"
)

func (svc *Service) newConcatenateFlow(jobID string, job *Job) (func() error, error) {
	zapFields := []zap.Field{zap.String("jobID", jobID), zap.Any("job", job)}
	type Params struct {
		Variants   []string `json:"variants"`
		AudioCodec string   `json:"audioCodec"`
		UploadURL  string   `json:"uploadUrl"`
	}
	params := Params{}
	err := mapToStruct(job.Params, &params)
	if err != nil {
		return nil, zaperr.Wrap(err, "failed to parse job params to %T: %w", zapFields...)
	}
	if params.AudioCodec == "" {
		params.AudioCodec = "copy"
	}
	zapFields = append(zapFields, zap.Any("params", params))
	svc.log.Debug("parsed job params", zapFields...)

	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		job, err := svc.storage.GetJob(ctx, jobID)
		if err != nil {
			return zaperr.Wrap(err, "failed to get job: %w", zapFields...)
		}
		zapFields = append(zapFields, zap.Any("job", job))

		updateJobStatus := func(status string) {
			job.DisplayStatus = status
			if err = svc.storage.SaveJob(ctx, job); err != nil {
				zapFields := append([]zap.Field{
					zaperr.ToField(err),
					zap.String("state", job.DisplayStatus),
				}, zapFields...)
				svc.log.Error("failed to save job state, proceeding", zapFields...)
			}
		}

		updateJobStatus(JobStatusDownloading)
		svc.log.Debug("starting download", zapFields...)
		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		filepathsMap, err := svc.downloader.Download(ctx, job.URL, params.Variants)
		if err != nil {
			return zaperr.Wrap(err, "failed to download variants", zapFields...)
		}

		var resultFilepath string
		if len(params.Variants) == 1 {
			resultFilepath = filepathsMap[params.Variants[0]]
		} else {
			updateJobStatus(JobStatusProcessing)
			// translate requested variants into actual fs filepaths while preserving order
			fsFilepaths := make([]string, 0, len(filepathsMap))
			for _, filepath := range params.Variants {
				fsFilepaths = append(fsFilepaths, filepathsMap[filepath])
			}

			svc.log.Debug("starting conversion", zapFields...)
			ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			resultFilepath, err = svc.mediaProcessor.Concatenate(ctx, fsFilepaths, params.AudioCodec)
			if err != nil {
				return zaperr.Wrap(err, "failed to concatenate files", zapFields...)
			}
		}
		zapFields = append(zapFields, zap.String("localFilename", resultFilepath))

		info, err := svc.mediaProcessor.GetInfo(ctx, resultFilepath)
		if err != nil {
			return zaperr.Wrap(err, "failed to get info about result file", zapFields...)
		}
		zapFields = append(zapFields, zap.Any("info", info))
		job.ResultMediaDuration = info.Duration
		job.ResultFileBytes = info.FileLenBytes
		//no need to save, as long as next line is status update // _ = svc.storage.SaveJob(ctx, job)

		updateJobStatus(JobStatusUploading)
		svc.log.Debug("starting upload", zapFields...)
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		err = svc.uploader.Upload(ctx, resultFilepath, params.UploadURL)
		if err != nil {
			return zaperr.Wrap(err, "failed to upload result", zapFields...)
		}

		updateJobStatus(JobStatusComplete)
		svc.log.Debug("job complete", zapFields...)
		return nil
	}, nil
}
