package server

import (
	"net/http"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/version"
)

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(version.VERSION))
}
