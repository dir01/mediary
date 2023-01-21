package service

import (
	"context"
	"time"

	"go.uber.org/zap"
)

func NewService(
	downloader Downloader,
	storage Storage,
	jobsQueue JobsQueue,
	mediaProcessor MediaProcessor,
	uploader Uploader,
	logger *zap.Logger,
) *Service {
	svc := &Service{
		downloader:     downloader,
		storage:        storage,
		jobsQueue:      jobsQueue,
		mediaProcessor: mediaProcessor,
		uploader:       uploader,
		log:            logger,
	}
	return svc
}

func (svc *Service) Start() {
	svc.jobsQueue.Subscribe(context.Background(), "process", svc.onPublishedJob)
	svc.jobsQueue.Run()
}

func (svc *Service) Stop() {
	svc.jobsQueue.Shutdown()
}

type Service struct {
	downloader     Downloader
	storage        Storage
	jobsQueue      JobsQueue
	mediaProcessor MediaProcessor
	uploader       Uploader
	log            *zap.Logger
}

//go:generate minimock -i Downloader -o ./mocks/downloader_mock.go -g
type Downloader interface {
	// AcceptsURL tells whether the downloader can handle the given URL.
	AcceptsURL(url string) bool

	// GetMetadata returns the metadata for the given URL.
	GetMetadata(ctx context.Context, url string) (*Metadata, error)

	// Download returns a mapping of url-local filenames to fs-local filenames.
	// like: {"chapter_01/01.mp3": "/tmp/downloads/url_xxx/chapter_01/01.mp3"}
	Download(ctx context.Context, url string, filepaths []string) (filepathsMap map[string]string, err error)
}

//go:generate minimock -i JobsQueue -o ./mocks/jobs_queue_mock.go -g
type JobsQueue interface {
	Publish(ctx context.Context, jobType string, payload any) error
	Subscribe(ctx context.Context, jobType string, f func(payloadBytes []byte) error)
	Shutdown()
	Run()
}

//go:generate minimock -i Storage -o ./mocks/storage_mock.go -g
type Storage interface {
	GetMetadata(ctx context.Context, url string) (*Metadata, error)
	SaveMetadata(ctx context.Context, metadata *Metadata) error
	GetJob(ctx context.Context, id string) (*Job, error)
	SaveJob(ctx context.Context, job *Job) error
}

//go:generate minimock -i MediaProcessor -o ./mocks/media_processor_mock.go -g
type MediaProcessor interface {
	Concatenate(ctx context.Context, filepaths []string, audioCodec string) (resultFilepath string, err error)
	GetInfo(ctx context.Context, filepath string) (info *MediaInfo, err error)
}

type MediaInfo struct {
	Duration     time.Duration
	FileLenBytes int64
}

//go:generate minimock -i Uploader -o ./mocks/uploader_mock.go -g
type Uploader interface {
	Upload(ctx context.Context, filepath string, url string) (err error)
}
