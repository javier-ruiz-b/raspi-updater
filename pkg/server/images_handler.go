package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/config"
)

type ImagesHandler struct {
	conf *config.ServerConfig
}

func (i *ImagesHandler) imageVersionHandler(w http.ResponseWriter, r *http.Request) {
	image, err := i.findImage(mux.Vars(r)["image"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	re := regexp.MustCompile(`.*_(.*)\.img.*`)
	versionMatch := re.FindStringSubmatch(image.Name())
	if len(versionMatch) != 2 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Couldn't get version for " + image.Name()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(versionMatch[1]))
}

func (i *ImagesHandler) imagePartitionTableHandler(w http.ResponseWriter, r *http.Request) {
	image, err := i.findImage(mux.Vars(r)["image"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	//TODO: handle compressed images
	file, err := os.Open(image.Name())
	//TODO: copy 512bytes to temp file and open with disk.NewDisk()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

//----

func (i *ImagesHandler) findImage(imageName string) (fs.FileInfo, error) {
	matches, err := filepath.Glob(*i.conf.ImagesDir + "/" + imageName + "*.img*")
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("No matching image found for %s", imageName)
	}

	if len(matches) != 1 {
		return nil, fmt.Errorf(fmt.Sprintf("There are %d images matching %s", len(matches), strings.Join(matches, " ")))
	}

	return os.Stat(matches[0])
}
