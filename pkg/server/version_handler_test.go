package server

import (
	"net/http"
	"testing"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
	"github.com/stretchr/testify/assert"
)

func TestGetsVersion(t *testing.T) {
	response := getRequest(t, "/version")

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, version.VERSION, response.Body.String())
}
