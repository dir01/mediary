package service

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/samber/oops"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (svc *Service) newConcatenateFlow(jobID string, job *Job) (func(ctx context.Context) error, error) {
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

	return func(jobCtx context.Context) error {
		jobCtx, span := otel.Tracer("github.com/dir01/mediary/service").Start(jobCtx, "service.ConcatenateFlow",
			trace.WithAttributes(
				attribute.String("job.id", jobID),
				attribute.Int("variants.count", len(params.Variants)),
				attribute.String("audio_codec", params.AudioCodec),
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
		logAttrs = append(logAttrs, slog.Any("job", job))
		errCtx = errCtx.With("job", job)

		updateJobStatus := func(status string) {
			job.DisplayStatus = status
			statusCtx, statusCancel := context.WithTimeout(jobCtx, 1*time.Second)
			defer statusCancel()
			if err = svc.storage.SaveJob(statusCtx, job); err != nil {
				attrs := append([]any{
					slog.Any("error", err),
					slog.String("state", job.DisplayStatus),
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
				attribute.Int("variants.count", len(params.Variants)),
			),
		)
		filepathsMap, err := svc.downloader.Download(downloadCtx, job.URL, params.Variants)
		if err != nil {
			downloadSpan.RecordError(err)
			downloadSpan.SetStatus(codes.Error, err.Error())
			downloadSpan.End()
			return errCtx.Wrapf(err, "failed to download variants")
		}
		downloadSpan.End()

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
				fileInfo, infoErr := svc.mediaProcessor.GetInfo(downloadCtx, fsFilepaths[i])
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
			concatCtx, concatCancel := context.WithTimeout(jobCtx, 30*time.Minute)
			defer concatCancel()

			concatCtx, concatSpan := otel.Tracer("github.com/dir01/mediary/service").Start(concatCtx, "service.Concatenate",
				trace.WithAttributes(
					attribute.String("job.id", jobID),
					attribute.Int("files.count", len(fsFilepaths)),
					attribute.String("audio_codec", params.AudioCodec),
				),
			)
			resultFilepath, err = svc.mediaProcessor.Concatenate(concatCtx, fsFilepaths, params.AudioCodec)
			if err != nil {
				concatSpan.RecordError(err)
				concatSpan.SetStatus(codes.Error, err.Error())
				concatSpan.End()
				return errCtx.Wrapf(err, "failed to concatenate files")
			}
			concatSpan.End()

			// write ID3 chapter tags into the concatenated file
			if len(chapters) > 0 {
				if chapErr := svc.mediaProcessor.AddChapterTags(concatCtx, resultFilepath, chapters); chapErr != nil {
					svc.log.Warn("failed to add chapter tags, proceeding without chapters",
						append(logAttrs, slog.Any("error", chapErr))...)
				}
			}
		}
		logAttrs = append(logAttrs, slog.String("localFilename", resultFilepath))
		errCtx = errCtx.With("localFilename", resultFilepath)

		info, err := svc.mediaProcessor.GetInfo(downloadCtx, resultFilepath)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return errCtx.Wrapf(err, "failed to get info about result file")
		}
		logAttrs = append(logAttrs, slog.Any("info", info))
		errCtx = errCtx.With("info", info)
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

		err = svc.uploader.Upload(uploadCtx, resultFilepath, params.UploadURL)
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
