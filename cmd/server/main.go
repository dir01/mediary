package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dir01/mediary/downloader"
	"github.com/dir01/mediary/downloader/torrent"
	"github.com/dir01/mediary/downloader/ytdlp"
	mediary_http "github.com/dir01/mediary/http"
	"github.com/dir01/mediary/media_processor"
	"github.com/dir01/mediary/otelsetup"
	"github.com/dir01/mediary/service"
	jobsqueue "github.com/dir01/mediary/service/jobs_queue"
	"github.com/dir01/mediary/storage"
	"github.com/dir01/mediary/uploader"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// region env vars
	sqliteDBPath := os.Getenv("SQLITE_DB_PATH")
	if sqliteDBPath == "" {
		log.Fatal("SQLITE_DB_PATH environment variable is required")
	}

	bindAddr := "0.0.0.0:8080"
	if _, ok := os.LookupEnv("BIND_ADDR"); ok {
		bindAddr = os.Getenv("BIND_ADDR")
	}

	var isDebug bool
	if val, exists := os.LookupEnv("DEBUG"); exists && val != "" && val != "0" && val != "false" {
		isDebug = true
	}
	// endregion

	var stderrHandler slog.Handler
	if isDebug {
		stderrHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		stderrHandler = slog.NewJSONHandler(os.Stderr, nil)
	}

	otelShutdown, err := otelsetup.Setup(context.Background())
	if err != nil {
		log.Fatalf("failed to setup opentelemetry: %v", err)
	}
	defer func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelShutdown(shutCtx); err != nil {
			log.Printf("error shutting down opentelemetry: %v", err)
		}
	}()

	logHandler := stderrHandler
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		logHandler = otelsetup.NewMultiHandler(stderrHandler, otelsetup.NewOTelSlogHandler("mediary"))
	}
	logger := slog.New(logHandler)

	// torrentDownloader downloads torrents
	torrentDownloader, err := torrent.New(os.TempDir(), logger, false)
	if err != nil {
		log.Fatalf("error creating torrent downloader: %v", err)
	}
	// TORRENT_STATIC_PEERS is a comma-separated list of "host:port" peers to add to every torrent
	staticPeers, err := torrent.ParseStaticPeers(os.Getenv("TORRENT_STATIC_PEERS"))
	if err != nil {
		log.Fatalf("error parsing TORRENT_STATIC_PEERS: %v", err)
	}
	torrentDownloader.AddStaticPeers(staticPeers)

	// ytdlDownloader downloads YouTube videos (potentially - everything that https://github.com/yt-dlp/yt-dlp  supports)
	ytdlDownloader, err := ytdlp.New(os.TempDir(), logger)
	if err != nil {
		log.Fatalf("error creating ytdl downloader: %v", err)
	}

	// dwn is a composite downloader: it can download anything, as long as one of its minions knows how to
	dwn := downloader.NewCompositeDownloader([]service.Downloader{torrentDownloader, ytdlDownloader})

	db, err := storage.OpenSQLiteDB(sqliteDBPath)
	if err != nil {
		log.Fatalf("error opening sqlite database: %v", err)
	}
	defer func() { _ = db.Close() }()

	queue, err := jobsqueue.NewSQLJobsQueue(db, logger)
	if err != nil {
		log.Fatalf("error initializing sql jobs queue: %v", err)
	}

	store, err := storage.NewSQLiteStorage(db)
	if err != nil {
		log.Fatalf("error initializing sqlite storage: %v", err)
	}

	mediaProc, err := media_processor.NewFFMpegMediaProcessor(logger)
	if err != nil {
		log.Fatalf("error initializing media processor: %v", err)
	}

	upl, err := uploader.New()
	if err != nil {
		log.Fatalf("error initializing uploader: %v", err)
	}

	// Kubernetes stops the pod with SIGTERM. Without this, Go ends the process at once,
	// and deferred functions do not run. Then the queue does not release the claims of
	// running jobs, and they wait for the claim to expire (the job timeout).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()

	svc := service.NewService(dwn, store, queue, mediaProc, upl, logger)
	svc.Start()

	server := &http.Server{Addr: bindAddr, Handler: mediary_http.PrepareHTTPServerMux(svc)}
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Starting to listen on %s", bindAddr)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Println("received shutdown signal, stopping")
	case err := <-serverErr:
		log.Printf("http server stopped: %v", err)
	}

	// Stop the jobs first: the queue cancels the running flows and releases their
	// claims, so that the next process takes the jobs at once. The pod has 30s
	// (terminationGracePeriodSeconds) for all steps.
	svc.Stop()

	shutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.Shutdown(shutCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error shutting down http server: %v", err)
	}
	log.Println("stopped")
}
