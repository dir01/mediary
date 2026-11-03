package torrent

import "context"

// NewWithMeterProvider lets tests read the metrics of a downloader.
var NewWithMeterProvider = newDownloader

// ResolveStaticPeers lets tests read the resolved static peer addresses.
func (td *Downloader) ResolveStaticPeers(ctx context.Context) []string {
	var addrs []string
	for _, p := range td.resolveStaticPeers(ctx) {
		addrs = append(addrs, p.Addr.String())
	}
	return addrs
}
