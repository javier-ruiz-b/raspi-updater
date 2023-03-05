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

func (hc *HandlerConfig) imageVersionHandler(w http.ResponseWriter, r *http.Request) (int, []byte) {
	imageName := mux.Vars(r)["id"]

	image, err := hc.imageDir.FindImage(imageName)
	if err != nil {
		return http.StatusNotFound, []byte(err.Error())
	}

	re := regexp.MustCompile(`.*_(.*)\.img.*`)
	versionMatch := re.FindStringSubmatch(image.Name())
	if len(versionMatch) != 2 {
		return http.StatusNotFound, []byte("Couldn't get version for " + image.Name())
	}

	return http.StatusOK, []byte(versionMatch[1])
}

func (hc *HandlerConfig) imagePartitionTableHandler(w http.ResponseWriter, r *http.Request) (int, []byte) {
	imageName := mux.Vars(r)["id"]

	disk, err := getDisk(*hc.imageDir, imageName)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	err = gob.NewEncoder(w).Encode(disk.GetPartitionTable())
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	return http.StatusOK, nil
}

func (hc *HandlerConfig) imageDownload(w http.ResponseWriter, r *http.Request) (int, []byte) {
	imageName := mux.Vars(r)["id"]
	compressor := mux.Vars(r)["compression"]
	partitionIndex, err := strconv.Atoi(mux.Vars(r)["partitionIndex"])
	if err != nil {
		return http.StatusBadRequest, []byte(err.Error())
	}

	disk, err := getDisk(*hc.imageDir, imageName)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	if partitionIndex >= len(disk.GetPartitionTable().Partitions) {
		return http.StatusBadRequest, []byte("Partition not found")
	}

	partition := disk.GetPartitionTable().Partitions[partitionIndex]

	image, err := hc.imageDir.FindImage(imageName)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	stream, err := image.OpenImage()
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}
	defer stream.Close()

	startOffset := int64(disk.GetPartitionTable().SectorSize) * int64(partition.Start)
	_, err = io.CopyN(ioutil.Discard, stream, startOffset)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	stdin, stdout, err := compressPipe(compressor)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
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
		return http.StatusInternalServerError, []byte(err.Error())
		// log.Print("Error compressing image stdout", err)
	}
	return http.StatusOK, nil
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

func getDisk(imageDir images.ImageDir, imageName string) (*disk.Disk, error) {
	image, err := imageDir.FindImage(imageName)
	if err != nil {
		return nil, err
	}

	disk, err := image.ReadDisk()
	if err != nil {
		return nil, err
	}

	return disk, err
}
