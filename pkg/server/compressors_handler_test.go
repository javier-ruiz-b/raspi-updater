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

	var decoders []string
	assert.Nil(t, gob.NewDecoder(response.Body).Decode(&decoders))
	assert.ElementsMatch(t, []string{"lz4", "xz"}, decoders)
}
