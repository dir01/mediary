package tests

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func NewDocsHelper(t *testing.T, mux *http.ServeMux, readmeLocation, blockStartMarker, blockEndMarker string) *docsHelper {
	return &docsHelper{t: t, mux: mux, readmeLocation: readmeLocation, blockStartMarker: blockStartMarker, blockEndMarker: blockEndMarker}
}

type docsHelper struct {
	mux              *http.ServeMux
	t                *testing.T
	readmeLocation   string
	blockStartMarker string
	blockEndMarker   string
	stringBits       []string
}

func (h *docsHelper) InsertText(text string, args ...interface{}) {
	h.stringBits = append(h.stringBits, fmt.Sprintf(text, args...)+"\n")
}

func (h *docsHelper) PerformRequestForDocs(
	method, url string, body io.ReadSeeker, expectedStatusCode int,
	responseHandler func(*httptest.ResponseRecorder),
) {
	resp := h.PerformRequest(method, url, body, expectedStatusCode, responseHandler)

	if resp == nil {
		return
	}

	var reqString string
	if body != nil {
		if _, err := body.Seek(0, io.SeekStart); err != nil {
			h.t.Fatalf("error seeking body: %v", err)
		}
		reqBodyBytes, err := io.ReadAll(body)
		if err != nil {
			h.t.Fatalf("error reading request body: %v", err)
			return
		}
		reqString = string(reqBodyBytes)
	}

	h.stringBits = append(h.stringBits, h.formatCall(method, url, reqString, resp.Body.String()))
}

func (h *docsHelper) PerformRequest(
	method string, url string, body io.Reader, expectedStatusCode int,
	responseHandler func(recorder *httptest.ResponseRecorder),
) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		h.t.Fatalf("error creating request: %v", err)
	}
	rr := httptest.NewRecorder()
	h.mux.ServeHTTP(rr, req)

	if status := rr.Code; status != expectedStatusCode {
		h.t.Errorf("handler returned wrong status code: got %v want %v\n%s", status, expectedStatusCode, rr.Body.String())
		return nil
	}

	if responseHandler != nil {
		responseHandler(rr)
	}

	return rr
}

func (h *docsHelper) Finish() {
	readmeBytes, err := os.ReadFile(h.readmeLocation)
	if err != nil {
		h.t.Fatalf("error reading file: %v", err)
	}

	textToInsert := strings.Join(h.stringBits, "\n")
	textToInsert = strings.ReplaceAll(textToInsert, "'''", "`")
	readmeString := h.insertBetweenBlockMarkers(string(readmeBytes), textToInsert)

	err = os.WriteFile(h.readmeLocation, []byte(readmeString), 0644)
	if err != nil {
		h.t.Fatalf("error writing file: %v", err)
	}
}

func (h *docsHelper) formatCall(method, url, body, respString string) string {
	callParts := []string{
		fmt.Sprintf("```\n$ curl -X %s '%s'", method, url),
	}
	if body != "" {
		callParts = append(callParts, fmt.Sprintf("--data-raw='%s'", body))
	}
	callParts = append(callParts, "\n", respString, "\n```\n\n")
	return strings.Join(callParts, "")
}

func (h *docsHelper) insertBetweenBlockMarkers(s, toInsert string) string {
	beforeMarker := s[:strings.Index(s, h.blockStartMarker)]
	afterMarker := s[strings.Index(s, h.blockEndMarker)+len(h.blockEndMarker):]
	s = beforeMarker + h.blockStartMarker + "\n" + toInsert + h.blockEndMarker + afterMarker
	return s
}
