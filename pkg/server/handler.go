package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/images"
)

const API_IMAGES_VERSION string = "/images/{id}/version"
const API_IMAGES_PARTITION_TABLE string = "/images/{id}/partitionTable"
const API_IMAGES_DOWNLOAD string = "/images/{id}/download/{partitionIndex}-{compression}"
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
		imageDir: images.NewImageDir(*options.ImagesDir),
	}

	serveMux.HandleFunc(API_VERSION, versionHandler)
	serveMux.HandleFunc(API_UPDATE, hc.updateHandler)
	serveMux.HandleFunc(API_IMAGES_VERSION, ih.imageVersionHandler)
	serveMux.HandleFunc(API_IMAGES_PARTITION_TABLE, ih.imagePartitionTableHandler)
	serveMux.HandleFunc(API_IMAGES_DOWNLOAD, ih.imageDownload)

	return serveMux
}
