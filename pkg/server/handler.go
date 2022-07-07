package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
)

type HandlerConfig struct {
	binariesDir string
}

func newHandler(options *config.ServerConfig) http.Handler {
	serveMux := mux.NewRouter()
	hc := &HandlerConfig{
		binariesDir: *options.UpdaterDir,
	}

	serveMux.HandleFunc("/version", versionHandler)
	serveMux.HandleFunc("/update/{filename}", hc.updateHandler)
	return serveMux
}
