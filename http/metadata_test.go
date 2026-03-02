package http

import (
	"net/http"
	"testing"
)

func TestExtractURLParam_MagnetWithTrackerAndName(t *testing.T) {
	// Callers pass the magnet URL unencoded as a GET query parameter.
	// The '&' chars inside the magnet are mis-parsed as query separators,
	// stripping the tracker (&tr=) and display name (&dn=) — which are
	// essential for finding peers without DHT.
	rawQuery := "url=magnet:?xt=urn:btih:0B1313000B0C900685793A9A50DA13D260246F4B&tr=http%3A%2F%2Fbt3.t-ru.org%2Fann%3Fmagnet&dn=Foo"
	req, _ := http.NewRequest(http.MethodGet, "/metadata?"+rawQuery, nil)

	got := extractURLParam(req)
	want := "magnet:?xt=urn:btih:0B1313000B0C900685793A9A50DA13D260246F4B&tr=http://bt3.t-ru.org/ann?magnet&dn=Foo"
	if got != want {
		t.Errorf("extractURLParam() =\n  %q\nwant\n  %q", got, want)
	}
}

func TestExtractURLParam_PlainURL(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/metadata?url=https://example.com/audio.mp3", nil)
	got := extractURLParam(req)
	if got != "https://example.com/audio.mp3" {
		t.Errorf("extractURLParam() = %q", got)
	}
}
