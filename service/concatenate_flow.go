package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func (svc *Service) newConcatenateFlow(jobID string, state *JobState) (func() error, error) {
	type Params struct {
		Filepaths  []string `json:"filepaths"`
		AudioCodec string   `json:"audioCodec"`
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

		svc.log.Debug("starting download", zap.String("jobID", jobID))
		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		filepathsMap, err := svc.downloader.Download(ctx, job.URL, params.Filepaths)
		if err != nil {
			svc.log.Error("failed to download files", zap.String("jobID", jobID), zap.Error(err))
			return fmt.Errorf("failed to download files: %w", err)
		}

		// translate requested relative filepaths into actual fs filepaths while preserving order
		fsFilepaths := make([]string, 0, len(filepathsMap))
		for _, filepath := range params.Filepaths {
			fsFilepaths = append(fsFilepaths, filepathsMap[filepath])
		}

		svc.log.Debug("starting conversion", zap.String("jobID", jobID))
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		resultFilepath, err := svc.mediaProcessor.Concatenate(ctx, fsFilepaths, params.AudioCodec)

		svc.log.Debug("starting upload", zap.String("jobID", jobID))
		fmt.Println(resultFilepath)
		return nil
	}, nil
}
