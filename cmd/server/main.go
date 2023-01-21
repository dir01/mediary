package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dir01/mediary/config"
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
	conf := config.New()
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	// torrentDownloader downloads torrents
	torrentDownloader, err := torrent.NewTorrentDownloader(os.TempDir(), logger)
	if err != nil {
		log.Fatalf("error creating torrent downloader: %v", err)
	}

	// dwn is a composite downloader: it can download anything, as long as one of its minions knows how to
	dwn := downloader.NewDownloader([]service.Downloader{torrentDownloader})

	// redisClient will be used both for storage and queue, mostly because I've found some cloud redis with a free tier
	opt, _ := redis.ParseURL(conf.MustGetRedisURL())
	redisClient := redis.NewClient(opt)
	defer func() { _ = redisClient.Close() }()

	queue, err := jobsqueue.NewRedisJobsQueue(redisClient, 10, "mediary", logger)
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
