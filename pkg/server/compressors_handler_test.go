package server

import (
	"encoding/gob"
	"net/http"
	"testing"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/compression"
	"github.com/stretchr/testify/assert"
)

func TestGetsCompressors(t *testing.T) {
	compression.SetupWindowsTests()
	response := getRequest(t, API_COMPRESSORS)
	assert.Equal(t, http.StatusOK, response.Code)

	var availableCompressors []string
	assert.Nil(t, gob.NewDecoder(response.Body).Decode(&availableCompressors))
	assert.Contains(t, availableCompressors, "lz4fast")
	assert.Contains(t, availableCompressors, "lz4best")
	assert.Contains(t, availableCompressors, "xz")
}
