package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func NewService(
	downloader Downloader,
	storage Storage,
	jobsQueue JobsQueue,
	mediaProcessor MediaProcessor,
	uploader Uploader,
	logger *slog.Logger,
) *Service {
	meter := otel.Meter("github.com/dir01/mediary/service")

	jobsCreated, err := meter.Int64Counter("mediary.jobs.created",
		metric.WithDescription("Total number of jobs created"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create mediary.jobs.created counter: %v", err))
	}

	jobsCompleted, err := meter.Int64Counter("mediary.jobs.completed",
		metric.WithDescription("Total number of jobs completed"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create mediary.jobs.completed counter: %v", err))
	}

	jobDuration, err := meter.Float64Histogram("mediary.jobs.duration_seconds",
		metric.WithDescription("Duration of job execution in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create mediary.jobs.duration_seconds histogram: %v", err))
	}

	svc := &Service{
		downloader:     downloader,
		storage:        storage,
		jobsQueue:      jobsQueue,
		mediaProcessor: mediaProcessor,
		uploader:       uploader,
		log:            logger,
		syncChansMap:   make(map[string]chan func()),
		jobsCreated:    jobsCreated,
		jobsCompleted:  jobsCompleted,
		jobDuration:    jobDuration,
	}
	return svc
}

func (svc *Service) Start() {
	svc.jobsQueue.Run()
	svc.jobsQueue.Subscribe(context.Background(), "process", svc.onPublishedJob)
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
	log            *slog.Logger

	// syncChansMap is used to synchronize the execution of the same task.
	// Map key is the task key.
	// Map value is the channel to which the task is sent.
	syncChansMap map[string]chan func()

	// syncChansMapMutex is used to synchronize access to syncChansMap.
	syncChansMapMutex sync.Mutex

	// OTel metric instruments
	jobsCreated   metric.Int64Counter
	jobsCompleted metric.Int64Counter
	jobDuration   metric.Float64Histogram
}

//go:generate  go tool github.com/gojuno/minimock/v3/cmd/minimock -i Downloader -o ./mocks/downloader_mock.go -g
type Downloader interface {
	// AcceptsURL tells whether the downloader can handle the given URL.
	AcceptsURL(url string) bool

	// GetMetadata returns the metadata for the given URL.
	GetMetadata(ctx context.Context, url string) (*Metadata, error)

	// Download returns a mapping of url-local filenames to fs-local filenames.
	// like: {"chapter_01/01.mp3": "/tmp/downloads/url_xxx/chapter_01/01.mp3"}
	Download(ctx context.Context, url string, filepaths []string) (filepathsMap map[string]string, err error)
}

//go:generate  go tool github.com/gojuno/minimock/v3/cmd/minimock -i JobsQueue -o ./mocks/jobs_queue_mock.go -g
type JobsQueue interface {
	Publish(ctx context.Context, jobType string, payload any) error
	Subscribe(ctx context.Context, jobType string, f func(ctx context.Context, payloadBytes []byte) error)
	Shutdown()
	Run()
}

//go:generate  go tool github.com/gojuno/minimock/v3/cmd/minimock -i Storage -o ./mocks/storage_mock.go -g
type Storage interface {
	GetMetadata(ctx context.Context, url string) (*Metadata, error)
	SaveMetadata(ctx context.Context, metadata *Metadata) error
	GetJob(ctx context.Context, id string) (*Job, error)
	SaveJob(ctx context.Context, job *Job) error
}

//go:generate  go tool github.com/gojuno/minimock/v3/cmd/minimock -i MediaProcessor -o ./mocks/media_processor_mock.go -g
type MediaProcessor interface {
	Concatenate(ctx context.Context, filepaths []string, audioCodec string) (resultFilepath string, err error)
	GetInfo(ctx context.Context, filepath string) (info *MediaInfo, err error)
	AddChapterTags(ctx context.Context, filepath string, chapters []Chapter) error
}

type MediaInfo struct {
	Duration     time.Duration
	FileLenBytes int64
}

type Chapter struct {
	Title     string
	StartTime time.Duration
	EndTime   time.Duration
}

//go:generate  go tool github.com/gojuno/minimock/v3/cmd/minimock -i Uploader -o ./mocks/uploader_mock.go -g
type Uploader interface {
	Upload(ctx context.Context, filepath string, url string) (err error)
}

// execSynced allows to execute the same task only once at a time, but only per-key.
// If the task is already running for the given key, the task will be queued and executed after the current task is finished.
// If another key is used, the task will be executed immediately.
func (svc *Service) execSynced(key string, f func()) {
	// region get or create channel
	svc.syncChansMapMutex.Lock()
	ch, ok := svc.syncChansMap[key]
	if !ok {
		ch = make(chan func(), 100) // 100 concurrent unique tasks is more than enough
		go func() {
			for fn := range ch {
				// execute the task
				fn()

				// after the task is executed, check if the channel is empty and close it if it is
				svc.syncChansMapMutex.Lock()
				if len(ch) == 0 {
					close(ch)
					delete(svc.syncChansMap, key)
				}
				svc.syncChansMapMutex.Unlock()
			}
		}()
		svc.syncChansMap[key] = ch
	}
	svc.syncChansMapMutex.Unlock()
	// endregion

	ch <- f
}
