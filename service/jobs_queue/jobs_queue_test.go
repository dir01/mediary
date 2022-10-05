package jobs_queue

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/tests"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()

func TestNewRedisJobsQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	redisURL, teardown, err := tests.GetFakeRedisURL(ctx)
	defer teardown()
	if err != nil {
		t.Fatalf("error getting redis url: %v", err)
	}

	opt, _ := redis.ParseURL(redisURL)
	redisClient := redis.NewClient(opt)
	defer func() { _ = redisClient.Close() }()

	t.Run("job is persisted", func(t *testing.T) {
		// First publish, then subscribe. Job should arrive
		queue, err := NewRedisJobsQueue(redisClient, logger, 10, randomPrefix())
		if err != nil {
			t.Errorf("error creating redis jobs queue: %v", err)
		}
		defer queue.Shutdown()

		job := &service.Job{ID: "some-id"}
		err = queue.Publish(ctx, job.ID)
		if err != nil {
			t.Errorf("error publishing job: %v", err)
		}

		var callCountMutex sync.RWMutex
		callCountMutex.Lock()
		callCount := 0
		callCountMutex.Unlock()
		queue.Subscribe(func(jobID string) error {
			callCountMutex.Lock()
			defer callCountMutex.Unlock()
			callCount++
			return nil
		})

		if eventually(2*time.Second, func() bool {
			callCountMutex.RLock()
			defer callCountMutex.RUnlock()
			return callCount == 1
		}) != true {
			t.Errorf("job was never delivered to subscriber")
		}
	})

	t.Run("job is retried", func(t *testing.T) {
		// If job is failed once, it will be retried shortly after
		queue, err := NewRedisJobsQueue(redisClient, logger, 10, randomPrefix())
		if err != nil {
			t.Errorf("error creating redis jobs queue: %v", err)
		}
		defer queue.Shutdown()

		job := &service.Job{ID: "some-id"}
		err = queue.Publish(ctx, job.ID)
		if err != nil {
			t.Errorf("error publishing job: %v", err)
		}

		var callCountMutex sync.RWMutex
		callCountMutex.Lock()
		callCount := 0
		callCountMutex.Unlock()
		queue.Subscribe(func(jobID string) error {
			callCountMutex.Lock()
			defer callCountMutex.Unlock()
			callCount++
			if callCount < 2 {
				return fmt.Errorf("some error")
			}
			return nil
		})

		if eventually(60*time.Second, func() bool {
			callCountMutex.RLock()
			defer callCountMutex.RUnlock()
			return callCount == 2
		}) != true {
			t.Errorf("job was never retried")
		}
	})
}

func eventually(timeout time.Duration, f func() bool) bool {
	timeoutChan := time.After(timeout)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-timeoutChan:
			return false
		case <-tick.C:
			if f() {
				return true
			}
		}
	}
}

func randomPrefix() (str string) {
	b := make([]byte, 24)
	rand.Read(b)
	return fmt.Sprintf("mediary:%x", b)

}
