package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
	"github.com/stretchr/testify/assert"
)

func TestGetsVersion(t *testing.T) {
	response := getRequest(t, "/version")

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, version.VERSION, response.Body.String())
}

func TestGets404OnUnexistingUrl(t *testing.T) {
	response := getRequest(t, "/unknown_url")

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestDownloadsUpdater(t *testing.T) {
	file := createTempFile(t)
	defer os.Remove(file.Name())
	response := getRequest(t, "/update/windows-amd64")
	assert.Equal(t, http.StatusOK, response.Code)

	_, err := io.Copy(file, response.Body)
	assert.Nil(t, err)

	fileContents, err := os.ReadFile(file.Name())
	assert.Nil(t, err)
	assert.Contains(t, string(fileContents), "Test file")
}

func createTempFile(t *testing.T) *os.File {
	file, err := os.CreateTemp(os.TempDir(), "updater_")
	assert.Nil(t, err)

	return file
}

func getRequest(t *testing.T, url string) *httptest.ResponseRecorder {
	options := config.NewServerConfig()
	binaryDir := "../testdata/bin"
	options.UpdaterDir = &binaryDir
	tested := newHandler(options)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	tested.ServeHTTP(response, req)
	return response
}
