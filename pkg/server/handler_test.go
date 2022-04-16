package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
	"github.com/stretchr/testify/assert"
)

func TestGetsVersion(t *testing.T) {
	tested := newHandler()

	req, err := http.NewRequest("GET", "/version", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	tested.ServeHTTP(rec, req)

	// Check the status code is what we expect.
	if status := rec.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, version.VERSION, rec.Body.String())
}
