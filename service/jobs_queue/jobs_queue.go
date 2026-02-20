package jobsqueue

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samber/oops"
	work2 "github.com/taylorchu/work"
)

type RJQ struct {
	work2Queue  work2.RedisQueue
	work2Worker *work2.Worker
	namespace   string
	concurrency int
	logger      *slog.Logger
}

func NewRedisJobsQueue(redisClient *redis.Client, concurrency int, namespace string, logger *slog.Logger) (*RJQ, error) {
	jobsQueue := &RJQ{
		work2Queue: work2.NewRedisQueue(redisClient),
		work2Worker: work2.NewWorker(&work2.WorkerOptions{
			Namespace: namespace,
			Queue:     work2.NewRedisQueue(redisClient),
			ErrorFunc: func(err error) {
				logger.Error("failed to handle job", slog.Any("error", err))
			},
		}),
		namespace:   namespace,
		concurrency: concurrency,
		logger:      logger,
	}
	return jobsQueue, nil
}

func (r *RJQ) Run() {
	r.work2Worker.Start()
}

func (r *RJQ) Shutdown() {
	r.work2Worker.Stop()
}

func (r *RJQ) Publish(ctx context.Context, jobType string, payload any) error {
	job := work2.NewJob()
	if err := job.MarshalJSONPayload(payload); err != nil {
		return oops.Wrapf(err, "failed to marshal payload")
	}

	if err := r.work2Queue.Enqueue(job, &work2.EnqueueOptions{Namespace: r.namespace, QueueID: jobType}); err != nil {
		return oops.Wrapf(err, "failed to enqueue job")
	}

	return nil
}

func (r *RJQ) Subscribe(ctx context.Context, jobType string, f func(payloadBytes []byte) error) {
	err := r.work2Worker.Register(jobType, func(job *work2.Job, opt *work2.DequeueOptions) error {
		// work silently ignores panics, so we need to recover them, log, and re-panic
		defer func() {
			if rec := recover(); rec != nil {
				r.logger.Error("panic recovered", slog.Any("panic", rec))
				panic(rec)
			}
		}()
		return f(job.Payload)
	}, &work2.JobOptions{
		MaxExecutionTime: 2 * time.Hour,
		IdleWait:         2 * time.Second,
		NumGoroutines:    int64(r.concurrency),
	})
	if err != nil {
		r.logger.Error("failed to register job", slog.Any("error", err))
	}
}
