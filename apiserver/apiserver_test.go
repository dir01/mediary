package apiserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dir01/mediary/apiserver"
	"github.com/stretchr/testify/assert"
)

func TestAPIServer_HandleHello(t *testing.T) {
	server := apiserver.New()
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	server.HandleHello().ServeHTTP(rec, req)
	assert.Equal(t, rec.Body.String(), "Hello world")
}
