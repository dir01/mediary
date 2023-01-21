package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dir01/mediary/downloader"
	"github.com/dir01/mediary/downloader/torrent"
	mediary_http "github.com/dir01/mediary/http"
	"github.com/dir01/mediary/media_processor"
	"github.com/dir01/mediary/service"
	"github.com/dir01/mediary/service/jobs_queue"
	"github.com/dir01/mediary/storage"
	"github.com/dir01/mediary/uploader"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	mustGetEnv := func(key string) string {
		value, ok := os.LookupEnv(key)
		if !ok {
			logger.Fatal("missing env var", zap.String("key", key))
		}
		return value
	}

	// torrentDownloader downloads torrents
	torrentDownloader, err := torrent.NewTorrentDownloader(os.TempDir(), logger)
	if err != nil {
		log.Fatalf("error creating torrent downloader: %v", err)
	}

	// dwn is a composite downloader: it can download anything, as long as one of its minions knows how to
	dwn := downloader.NewDownloader([]service.Downloader{torrentDownloader})

	mkRedisClient := func(url string) (client *redis.Client, teardown func()) {
		opt, err := redis.ParseURL(url)
		if err != nil {
			logger.Fatal("error parsing redis url", zap.Error(err))
		}
		redisClient := redis.NewClient(opt)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := redisClient.Ping(ctx).Result(); err != nil {
			logger.Fatal("error connecting to redis", zap.Error(err))
		}
		return redisClient, func() { _ = redisClient.Close() }
	}

	// redisClient will be used both for storage
	redisURL := mustGetEnv("REDIS_URL")
	redisClient, teardownRedis := mkRedisClient(redisURL)
	defer teardownRedis()

	// and for jobs queue
	bgRedisURL := os.Getenv("REDIS_URL_BG_JOBS")
	if bgRedisURL == "" {
		bgRedisURL = redisURL
	}
	bgRedisClient, teardownBgRedis := mkRedisClient(bgRedisURL)
	defer teardownBgRedis()

	queue, err := jobsqueue.NewRedisJobsQueue(bgRedisClient, 2, "mediary", logger)
	if err != nil {
		log.Fatalf("error initializing redis jobs queue: %v", err)
	}
	defer queue.Shutdown()

	store := storage.NewRedisStorage(redisClient, "mediary")

	mediaProc, err := media_processor.NewFFMpegMediaProcessor(logger)
	if err != nil {
		log.Fatalf("error initializing media processor: %v", err)
	}

	upl, err := uploader.New()
	if err != nil {
		log.Fatalf("error initializing uploader: %v", err)
	}

	svc := service.NewService(dwn, store, queue, mediaProc, upl, logger)
	svc.Start()
	defer svc.Stop()

	mux := mediary_http.PrepareHTTPServerMux(svc)

	addr := "0.0.0.0:8080"
	log.Printf("Starting to listen on %s", addr)
	log.Println(http.ListenAndServe(addr, mux))
}
