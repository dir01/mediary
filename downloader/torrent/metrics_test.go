package torrent_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	torrentdownloader "github.com/dir01/mediary/downloader/torrent"
	"github.com/dir01/mediary/tests"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestMetrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	dwn, err := torrentdownloader.NewWithMeterProvider(t.TempDir(), logger, false, sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}
	piece := bytes.Repeat([]byte{1}, tests.SeederPieceLength)
	seeder, magnetURL := tests.SetupPartialSeeder(t, "partial-collection", map[string][]byte{
		"01.mp3": piece,
		"02.mp3": piece,
		"03.mp3": piece,
	}, []string{"03.mp3"})
	dwn.AddBootstrapPeer(seeder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		// "03.mp3" has no seeder, thus the download waits until cancel.
		_, _ = dwn.Download(ctx, magnetURL, []string{"01.mp3", "02.mp3", "03.mp3"})
	}()

	name := attribute.String("torrent.name", "partial-collection")
	deadline := time.Now().Add(30 * time.Second)
	for {
		m := collect(t, reader)
		if v, ok := value(m, "mediary.download.bytes_completed", name); ok && v == 2*tests.SeederPieceLength {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("download progress was not reported: %v", m)
		}
		time.Sleep(100 * time.Millisecond)
	}

	m := collect(t, reader)
	expect(t, m, "mediary.download.bytes_total", 3*tests.SeederPieceLength, name)
	expectAtLeast(t, m, "mediary.download.bytes_read", 2*tests.SeederPieceLength, name)
	expectAtLeast(t, m, "mediary.torrent.bytes_read", 2*tests.SeederPieceLength)
	expect(t, m, "mediary.torrents", 1, attribute.String("torrent.state", "downloading"))
	expect(t, m, "mediary.torrents", 0, attribute.String("torrent.state", "idle"))
	expect(t, m, "mediary.torrents", 0, attribute.String("torrent.state", "metadata"))
	expectAtLeast(t, m, "mediary.download.peers", 1, name, attribute.String("peer.kind", "active"))
	expectAtLeast(t, m, "mediary.torrent.peers", 1, attribute.String("peer.kind", "active"))

	cancel()
	<-done

	m = collect(t, reader)
	expect(t, m, "mediary.torrents", 0, attribute.String("torrent.state", "downloading"))
	expect(t, m, "mediary.torrents", 1, attribute.String("torrent.state", "idle"))
	if _, ok := value(m, "mediary.download.bytes_completed", name); ok {
		t.Errorf("expected no download series after Download returns")
	}
}

type observations map[string][]metricdata.DataPoint[int64]

func collect(t *testing.T, reader *sdkmetric.ManualReader) observations {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}
	result := observations{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch data := m.Data.(type) {
			case metricdata.Sum[int64]:
				result[m.Name] = data.DataPoints
			case metricdata.Gauge[int64]:
				result[m.Name] = data.DataPoints
			}
		}
	}
	return result
}

func value(m observations, metricName string, attrs ...attribute.KeyValue) (int64, bool) {
	want := attribute.NewSet(attrs...)
	for _, dp := range m[metricName] {
		if dp.Attributes.Equals(&want) {
			return dp.Value, true
		}
	}
	return 0, false
}

func expect(t *testing.T, m observations, metricName string, want int64, attrs ...attribute.KeyValue) {
	t.Helper()
	if got, ok := value(m, metricName, attrs...); !ok || got != want {
		t.Errorf("%s%v: expected %d, got %d (found: %v)", metricName, attrs, want, got, ok)
	}
}

func expectAtLeast(t *testing.T, m observations, metricName string, want int64, attrs ...attribute.KeyValue) {
	t.Helper()
	if got, ok := value(m, metricName, attrs...); !ok || got < want {
		t.Errorf("%s%v: expected at least %d, got %d (found: %v)", metricName, attrs, want, got, ok)
	}
}
