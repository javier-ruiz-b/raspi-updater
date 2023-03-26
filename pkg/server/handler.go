package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/images"
)

const API_IMAGES_BACKUP string = "/images/{id}/backup/{version}/{compression}"
const API_IMAGES_BACKUP_EXISTS string = "/images/{id}/backup_exists/{version}"
const API_IMAGES_VERSION string = "/images/{id}/version"
const API_IMAGES_PARTITION_TABLE string = "/images/{id}/partitionTable"
const API_IMAGES_DOWNLOAD string = "/images/{id}/download/{partitionIndex}-{compression}"
const API_UPDATE string = "/update/{filename}"
const API_VERSION string = "/version"

type HandlerConfig struct {
	binariesDir string
	imageDir    *images.ImageDir
}

var verbose = false

func newMainHandler(options *config.ServerConfig) http.Handler {
	serveMux := mux.NewRouter()
	hc := &HandlerConfig{
		binariesDir: *options.UpdaterDir,
		imageDir:    images.NewImageDir(*options.ImagesDir),
	}
	verbose = *options.Verbose

	serveMux.Handle(API_VERSION, newPathHandler(hc.versionHandler))
	serveMux.Handle(API_UPDATE, newPathHandler(hc.updateHandler))
	serveMux.Handle(API_IMAGES_VERSION, newPathHandler(hc.imageVersionHandler))
	serveMux.Handle(API_IMAGES_PARTITION_TABLE, newPathHandler(hc.imagePartitionTableHandler))
	serveMux.Handle(API_IMAGES_DOWNLOAD, newPathHandler(hc.imageDownload))
	serveMux.Handle(API_IMAGES_BACKUP, newPathHandler(hc.imageBackup))
	serveMux.Handle(API_IMAGES_BACKUP_EXISTS, newPathHandler(hc.imageBackupExists))

	return serveMux
}

type PathHandler struct {
	handleFunc func(w http.ResponseWriter, r *http.Request) (int, []byte)
}

func newPathHandler(handleFunc func(w http.ResponseWriter, r *http.Request) (int, []byte)) *PathHandler {
	return &PathHandler{
		handleFunc: handleFunc,
	}
}

func (p *PathHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if verbose {
		log.Printf("%s: %s %s Request\n", r.Host, r.Method, r.URL)
	}

	statusCode, response := p.handleFunc(w, r)

	if statusCode != http.StatusOK {
		log.Printf("%s: %s %s Error [%d] %s\n", r.Host, r.Method, r.URL, statusCode, string(response))
	} else if verbose {
		log.Printf("%s: %s %s OK\n", r.Host, r.Method, r.URL)
	}

	w.WriteHeader(statusCode)
	if response != nil {
		w.Write(response)
	}
}
