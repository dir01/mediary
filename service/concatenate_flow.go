package service

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/samber/oops"
)

func (svc *Service) newConcatenateFlow(jobID string, job *Job) (func() error, error) {
	logAttrs := []any{slog.String("jobID", jobID), slog.Any("job", job)}
	errCtx := oops.With("jobID", jobID, "job", job)
	type Params struct {
		Variants   []string `json:"variants"`
		AudioCodec string   `json:"audioCodec"`
		UploadURL  string   `json:"uploadUrl"`
	}
	params := Params{}
	err := mapToStruct(job.Params, &params)
	if err != nil {
		return nil, errCtx.Wrapf(err, "failed to parse job params")
	}
	if params.AudioCodec == "" {
		params.AudioCodec = "copy"
	}
	logAttrs = append(logAttrs, slog.Any("params", params))
	errCtx = errCtx.With("params", params)
	svc.log.Debug("parsed job params", logAttrs...)

	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		job, err := svc.storage.GetJob(ctx, jobID)
		if err != nil {
			return errCtx.Wrapf(err, "failed to get job")
		}
		logAttrs = append(logAttrs, slog.Any("job", job))
		errCtx = errCtx.With("job", job)

		updateJobStatus := func(status string) {
			job.DisplayStatus = status
			if err = svc.storage.SaveJob(ctx, job); err != nil {
				attrs := append([]any{
					slog.Any("error", err),
					slog.String("state", job.DisplayStatus),
				}, logAttrs...)
				svc.log.Error("failed to save job state, proceeding", attrs...)
			}
		}

		updateJobStatus(JobStatusDownloading)
		svc.log.Debug("starting download", logAttrs...)
		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		filepathsMap, err := svc.downloader.Download(ctx, job.URL, params.Variants)
		if err != nil {
			return errCtx.Wrapf(err, "failed to download variants")
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
						append(logAttrs, slog.Any("error", infoErr), slog.String("variant", variant))...)
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

			svc.log.Debug("starting conversion", logAttrs...)
			ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			resultFilepath, err = svc.mediaProcessor.Concatenate(ctx, fsFilepaths, params.AudioCodec)
			if err != nil {
				return errCtx.Wrapf(err, "failed to concatenate files")
			}

			// write ID3 chapter tags into the concatenated file
			if len(chapters) > 0 {
				if chapErr := svc.mediaProcessor.AddChapterTags(ctx, resultFilepath, chapters); chapErr != nil {
					svc.log.Warn("failed to add chapter tags, proceeding without chapters",
						append(logAttrs, slog.Any("error", chapErr))...)
				}
			}
		}
		logAttrs = append(logAttrs, slog.String("localFilename", resultFilepath))
		errCtx = errCtx.With("localFilename", resultFilepath)

		info, err := svc.mediaProcessor.GetInfo(ctx, resultFilepath)
		if err != nil {
			return errCtx.Wrapf(err, "failed to get info about result file")
		}
		logAttrs = append(logAttrs, slog.Any("info", info))
		errCtx = errCtx.With("info", info)
		job.ResultMediaDuration = info.Duration
		job.ResultFileBytes = info.FileLenBytes
		//no need to save, as long as next line is status update // _ = svc.storage.SaveJob(ctx, job)

		updateJobStatus(JobStatusUploading)
		svc.log.Debug("starting upload", logAttrs...)
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		err = svc.uploader.Upload(ctx, resultFilepath, params.UploadURL)
		if err != nil {
			return errCtx.Wrapf(err, "failed to upload result")
		}

		updateJobStatus(JobStatusComplete)
		svc.log.Debug("job complete", logAttrs...)
		return nil
	}, nil
}
