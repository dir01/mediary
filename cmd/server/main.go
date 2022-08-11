package main

import (
	"log"
	"net/http"

	"github.com/dir01/mediary"
)

func main() {
	torrentDownloader, err := mediary.NewTorrentDownloader()
	if err != nil {
		log.Fatalf("error creating torrent downloader: %w", err)
	}
	service := mediary.NewService([]mediary.Downloader{torrentDownloader}, mediary.NewStorageInMemory())
	mux := mediary.PrepareHTTPServerMux(service)
	addr := "0.0.0.0:8080"
	log.Printf("Starting to listen on %s", addr)
	log.Println(http.ListenAndServe(addr, mux))
}
