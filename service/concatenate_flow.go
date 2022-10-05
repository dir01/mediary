package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func (svc *Service) newConcatenateFlow(jobID string, state *Job) (func() error, error) {
	type Params struct {
		Filepaths  []string `json:"filepaths"`
		AudioCodec string   `json:"audioCodec"`
		UploadURL  string   `json:"uploadUrl"`
	}
	params := Params{}
	err := mapToStruct(state.Params, &params)
	if err != nil {
		svc.log.Error("failed to parse job params", zap.String("jobID", jobID), zap.Error(err))
		return nil, fmt.Errorf("failed to parse job params to %T: %w", params, err)
	}
	if params.AudioCodec == "" {
		params.AudioCodec = "copy"
	}
	svc.log.Debug("parsed job params", zap.String("jobID", jobID), zap.Any("params", params))

	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		job, err := svc.storage.GetJob(ctx, jobID)
		if err != nil {
			svc.log.Error("failed to get job state", zap.String("jobID", jobID), zap.Error(err))
			return fmt.Errorf("failed to get job: %w", err)
		}

		updateJobStatus := func(status string) {
			job.DisplayStatus = status
			if err = svc.storage.SaveJob(ctx, job); err != nil {
				svc.log.Error(
					"failed to save job state, proceeding",
					zap.String("jobID", jobID), zap.String("state", job.DisplayStatus), zap.Error(err),
				)
			}
		}

		updateJobStatus(StatusDownloading)
		svc.log.Debug("starting download", zap.String("jobID", jobID))
		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		filepathsMap, err := svc.downloader.Download(ctx, job.URL, params.Filepaths)
		if err != nil {
			svc.log.Error("failed to download files", zap.String("jobID", jobID), zap.Error(err))
			return fmt.Errorf("failed to download files: %w", err)
		}

		updateJobStatus(StatusProcessing)
		// translate requested relative filepaths into actual fs filepaths while preserving order
		fsFilepaths := make([]string, 0, len(filepathsMap))
		for _, filepath := range params.Filepaths {
			fsFilepaths = append(fsFilepaths, filepathsMap[filepath])
		}

		svc.log.Debug("starting conversion", zap.String("jobID", jobID))
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		resultFilepath, err := svc.mediaProcessor.Concatenate(ctx, fsFilepaths, params.AudioCodec)
		if err != nil {
			svc.log.Error("failed to concatenate files", zap.String("jobID", jobID), zap.Error(err))
			return fmt.Errorf("failed to concatenate files: %w", err)
		}

		updateJobStatus(StatusUploading)
		svc.log.Debug("starting upload", zap.String("jobID", jobID))
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		err = svc.uploader.Upload(ctx, resultFilepath, params.UploadURL)
		if err != nil {
			svc.log.Error(
				"failed to upload result",
				zap.String("jobID", jobID),
				zap.String("localFilename", resultFilepath),
				zap.Error(err),
			)
			return fmt.Errorf("failed to upload result: %w", err)
		}

		updateJobStatus(StatusComplete)
		svc.log.Debug("job complete", zap.String("jobID", jobID))
		return nil
	}, nil
}
