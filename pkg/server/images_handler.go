package server

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/images"
)

type ImagesHandler struct {
	imageDir *images.ImageDir
}

func (i *ImagesHandler) imageVersionHandler(w http.ResponseWriter, r *http.Request) {
	imageName := mux.Vars(r)["image"]
	image, err := i.imageDir.FindImage(imageName)
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
	imageName := mux.Vars(r)["image"]
	image, err := i.imageDir.FindImage(imageName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	disk, err := image.ReadDisk()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	var buffer bytes.Buffer
	err = gob.NewEncoder(&buffer).Encode(disk)
	if err != nil {
		//TODO: disk has no exported fields.
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(buffer.Bytes())
}

//----
