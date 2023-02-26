package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
)

const API_IMAGES_VERSION string = "/images/{image}/version"
const API_IMAGES_PARTITION_TABLE string = "/images/{image}/partitionTable"
const API_UPDATE string = "/update/{filename}"
const API_VERSION string = "/version"

type HandlerConfig struct {
	binariesDir string
}

func newHandler(options *config.ServerConfig) http.Handler {
	serveMux := mux.NewRouter()
	hc := &HandlerConfig{
		binariesDir: *options.UpdaterDir,
	}
	ih := &ImagesHandler{
		conf: options,
	}

	serveMux.HandleFunc(API_VERSION, versionHandler)
	serveMux.HandleFunc(API_UPDATE, hc.updateHandler)
	serveMux.HandleFunc(API_IMAGES_VERSION, ih.imageVersionHandler)
	serveMux.HandleFunc(API_IMAGES_PARTITION_TABLE, ih.imagePartitionTableHandler)
	return serveMux
}
