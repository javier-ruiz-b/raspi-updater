package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type HandlerConfig struct {
	binariesDir string
}

func newHandler(binariesDir string) http.Handler {
	serveMux := mux.NewRouter()
	hc := &HandlerConfig{
		binariesDir: binariesDir,
	}

	serveMux.HandleFunc("/version", versionHandler)
	serveMux.HandleFunc("/update/{filename}", hc.updateHandler)
	return serveMux
}
