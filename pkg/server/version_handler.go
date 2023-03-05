package server

import (
	"net/http"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
)

func (hc *HandlerConfig) versionHandler(w http.ResponseWriter, r *http.Request) (int, []byte) {
	return http.StatusOK, []byte(version.VERSION)
}
