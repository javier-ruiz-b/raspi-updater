package server

import (
	"encoding/gob"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/disk"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/images"
)

type ImagesHandler struct {
	imageDir *images.ImageDir
}

func (i *ImagesHandler) imageVersionHandler(w http.ResponseWriter, r *http.Request) {
	imageName := mux.Vars(r)["id"]
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
	imageName := mux.Vars(r)["id"]

	disk, err := i.getDisk(imageName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = gob.NewEncoder(w).Encode(disk.GetPartitionTable())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (i *ImagesHandler) imageDownload(w http.ResponseWriter, r *http.Request) {
	imageName := mux.Vars(r)["id"]
	compressor := mux.Vars(r)["compression"]
	partitionIndex, err := strconv.Atoi(mux.Vars(r)["partitionIndex"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	disk, err := i.getDisk(imageName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if partitionIndex >= len(disk.GetPartitionTable().Partitions) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Partition not found"))
		return
	}

	partition := disk.GetPartitionTable().Partitions[partitionIndex]

	image, err := i.imageDir.FindImage(imageName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	stream, err := image.OpenImage()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer stream.Close()

	startOffset := int64(disk.GetPartitionTable().SectorSize) * int64(partition.Start)
	_, err = io.CopyN(ioutil.Discard, stream, startOffset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	stdin, stdout, err := compressPipe(compressor)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)

	size := int64(partition.Size) * int64(disk.GetPartitionTable().SectorSize)
	go func() {
		_, err := io.CopyN(stdin, stream, size)
		stdin.Close()
		if err != nil {
			log.Print("Error compressing image stdin", err)
		}
		// log.Print("Raw ", passedBytes, " bytes")
	}()
	_, err = io.Copy(w, stdout)
	if err != nil {
		log.Print("Error compressing image stdout", err)
	}
	// log.Print("Compressed ", num, " bytes")
}

// -----

func compressPipe(compressor string) (io.WriteCloser, io.ReadCloser, error) {
	cmd := exec.Command(compressor, "-c", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	go cmd.Run()

	return stdin, stdout, err
}

func (i *ImagesHandler) getDisk(imageName string) (*disk.Disk, error) {
	image, err := i.imageDir.FindImage(imageName)
	if err != nil {
		return nil, err
	}

	disk, err := image.ReadDisk()
	if err != nil {
		return nil, err
	}

	return disk, err
}
