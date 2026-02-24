package service

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/hori-ryota/zaperr"
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
			for _, fp := range params.Variants {
				fsFilepaths = append(fsFilepaths, filepathsMap[fp])
			}

			// collect per-file durations for chapter markers
			var chapters []Chapter
			var offset time.Duration
			for i, variant := range params.Variants {
				fileInfo, infoErr := svc.mediaProcessor.GetInfo(ctx, fsFilepaths[i])
				if infoErr != nil {
					svc.log.Warn("failed to get duration for chapter, skipping chapter tags",
						append(zapFields, zaperr.ToField(infoErr), zap.String("variant", variant))...)
					chapters = nil
					break
				}
				if fileInfo.Duration <= 0 {
					svc.log.Warn("zero or negative duration for chapter, skipping chapter tags",
						append(zapFields, zap.Duration("duration", fileInfo.Duration), zap.String("variant", variant))...)
					chapters = nil
					break
				}
				name := strings.TrimSuffix(filepath.Base(variant), filepath.Ext(variant))
				chapters = append(chapters, Chapter{
					Title:     name,
					StartTime: offset,
					EndTime:   offset + fileInfo.Duration,
				})
				offset += fileInfo.Duration
			}

			svc.log.Debug("starting conversion", zapFields...)
			ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			resultFilepath, err = svc.mediaProcessor.Concatenate(ctx, fsFilepaths, params.AudioCodec)
			if err != nil {
				return zaperr.Wrap(err, "failed to concatenate files", zapFields...)
			}

			// write ID3 chapter tags into the concatenated file
			if len(chapters) > 0 {
				if chapErr := svc.mediaProcessor.AddChapterTags(ctx, resultFilepath, chapters); chapErr != nil {
					svc.log.Warn("failed to add chapter tags, proceeding without chapters",
						append(zapFields, zaperr.ToField(chapErr))...)
				}
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
