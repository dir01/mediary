package downloader

import (
	"context"
	"fmt"

	"github.com/dir01/mediary/service"
)

var ErrUrlNotSupported = fmt.Errorf("url not supported")

func NewDownloader(downloaders []service.Downloader) *Downloader {
	downloader := &Downloader{downloaders}
	var _ service.Downloader = downloader
	return downloader
}

type Downloader struct {
	downloaders []service.Downloader
}

func (d *Downloader) AcceptsURL(url string) bool {
	if d.getConcreteDownloader(url) == nil {
		return false
	} else {
		return true
	}
}

func (d *Downloader) GetMetadata(ctx context.Context, url string) (*service.Metadata, error) {
	if downloader := d.getConcreteDownloader(url); downloader == nil {
		return nil, ErrUrlNotSupported
	} else {
		return downloader.GetMetadata(ctx, url)
	}
}

func (d *Downloader) Download(ctx context.Context, url string, filepaths []string) (filepathsMap map[string]string, err error) {
	if downloader := d.getConcreteDownloader(url); downloader == nil {
		return nil, ErrUrlNotSupported
	} else {
		return downloader.Download(ctx, url, filepaths)
	}
}

func (d *Downloader) getConcreteDownloader(url string) service.Downloader {
	for _, downloader := range d.downloaders {
		if downloader.AcceptsURL(url) {
			return downloader
		}
	}
	return nil
}
