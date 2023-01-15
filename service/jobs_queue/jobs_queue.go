package jobs_queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dir01/mediary/service"
	"github.com/go-redis/redis"
	"github.com/robinjoseph08/redisqueue"
	"go.uber.org/zap"
)

func NewRedisJobsQueue(redisClient *redis.Client, logger *zap.Logger, concurrency int, keyPrefix string) (*RedisJobsQueue, error) {
	p, err := redisqueue.NewProducerWithOptions(&redisqueue.ProducerOptions{
		StreamMaxLength:      100,
		ApproximateMaxLength: true,
		RedisClient:          redisClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create redisqueue producer: %w", err)
	}

	c, err := redisqueue.NewConsumerWithOptions(&redisqueue.ConsumerOptions{
		RedisClient: redisClient,
		// BlockingTimeout says for how long can we block for a message to be available.
		// If there are no new messages, this is how long we'll wait before a graceful shutdown.
		BlockingTimeout: 1 * time.Second,
		// Concurrency sets the number of goroutines spawned to consume messages.
		// This effectively sets how many jobs can be processed at the same time
		Concurrency: concurrency,
		// VisibilityTimeout sets how long a message is invisible to other consumers
		// so if a consumer dies and never comes back, after this timeout it will be available for other consumers
		VisibilityTimeout: 8 * time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create redisqueue consumer: %w", err)
	}

	go func() { // consumer errors must be consumed or else the consumer will block on error
		for {
			select {
			case err := <-c.Errors:
				logger.Error("redisqueue consumer error", zap.Error(err))
			}
		}
	}()

	streamName := fmt.Sprintf("%s:%s", keyPrefix, "jobs")

	r := &RedisJobsQueue{
		producer:   p,
		consumer:   c,
		streamName: streamName,
		subscribed: make(chan struct{}),
	}

	go func() {
		<-r.subscribed // consumer.Run() can not be run until we have at least one consumer
		c.Run()        // consumer.Run() blocks, so we need to run it in a separate goroutine
	}()

	var _ service.JobsQueue = r
	return r, nil
}

type RedisJobsQueue struct {
	producer   *redisqueue.Producer
	consumer   *redisqueue.Consumer
	streamName string
	subscribed chan struct{}
	subOnce    sync.Once
}

func (r *RedisJobsQueue) Publish(ctx context.Context, jobId string) error {
	err := r.producer.Enqueue(&redisqueue.Message{
		Stream: r.streamName,
		Values: map[string]interface{}{"jobID": jobId},
	})
	if err != nil {
		return fmt.Errorf("failed to publish job: %w", err)
	}
	return nil
}

func (r *RedisJobsQueue) Subscribe(f func(jobId string) error) {
	r.consumer.Register(r.streamName, func(msg *redisqueue.Message) error {
		return retry(func() error {
			// redisqueue does not seem to care about retries, so first we'll try to retry in-process
			// interestingly, on server restart unacked messages will be re-scheduled, so we don't need to worry about that
			// also, in case we die completely and consumer on another host will be started with another name,
			// it will pick this message after `VisibilityTimeout` and will be able to process it
			jobId, ok := msg.Values["jobID"].(string)
			if !ok {
				return fmt.Errorf("invalid jobID")
			}
			return f(jobId)
		}, 1*time.Second, 1*time.Minute, 5*time.Minute, 10*time.Minute, 30*time.Minute, 1*time.Hour, 2*time.Hour, 4*time.Hour)
	})
	r.subOnce.Do(func() {
		close(r.subscribed)
	})
}

func (r *RedisJobsQueue) Shutdown() {
	r.consumer.Shutdown()
}

func retry(fn func() error, durations ...time.Duration) error {
	var err error
	for _, dur := range durations {
		if err = fn(); err == nil {
			return nil
		} else {
			time.Sleep(dur)
			continue
		}
	}
	return err
}
