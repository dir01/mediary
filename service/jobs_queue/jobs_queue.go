package jobs_queue

import (
	"context"
	"fmt"
	"time"

	"github.com/dir01/mediary/service"
	"github.com/go-redis/redis"
	"github.com/robinjoseph08/redisqueue"
)

func NewRedisJobsQueue(redisClient *redis.Client, concurrency int, keyPrefix string) (*RedisJobsQueue, error) {
	p, err := redisqueue.NewProducerWithOptions(&redisqueue.ProducerOptions{
		StreamMaxLength:      1000,
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
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create redisqueue consumer: %w", err)
	}

	streamName := fmt.Sprintf("%s:%s", keyPrefix, "jobs")

	r := &RedisJobsQueue{producer: p, consumer: c, streamName: streamName}
	var _ service.JobsQueue = r
	return r, nil
}

type RedisJobsQueue struct {
	producer   *redisqueue.Producer
	consumer   *redisqueue.Consumer
	streamName string
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

func (r RedisJobsQueue) Subscribe(f func(jobId string) error) {
	r.consumer.Register(r.streamName, func(msg *redisqueue.Message) error {
		jobId, ok := msg.Values["jobID"].(string)
		if !ok {
			return fmt.Errorf("invalid jobID")
		}
		return f(jobId)
	})
	go r.consumer.Run() // FIXME
}
