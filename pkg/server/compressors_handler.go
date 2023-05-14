package server

import (
	"encoding/gob"
	"net/http"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/compression"
)

func (hc *HandlerConfig) compressorsHandler(w http.ResponseWriter, r *http.Request) (int, []byte) {
	availableCompressors := compression.AvailableToolMap()

	var compressorNames []string
	for name := range availableCompressors {
		compressorNames = append(compressorNames, name)
	}

	err := gob.NewEncoder(w).Encode(compressorNames)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	return http.StatusOK, nil
}
