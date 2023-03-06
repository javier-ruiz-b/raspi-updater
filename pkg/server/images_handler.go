package server

import (
	"encoding/gob"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/compression"
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

	partitionTable, err := getPartitionTable(*hc.imageDir, imageName)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	err = gob.NewEncoder(w).Encode(partitionTable)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	return http.StatusOK, nil
}

func (hc *HandlerConfig) imageDownload(w http.ResponseWriter, r *http.Request) (int, []byte) {
	imageName := mux.Vars(r)["id"]
	compressionBinary := mux.Vars(r)["compression"]
	partitionIndex, err := strconv.Atoi(mux.Vars(r)["partitionIndex"])
	if err != nil {
		return http.StatusBadRequest, []byte(err.Error())
	}

	partitionTable, err := getPartitionTable(*hc.imageDir, imageName)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	if partitionIndex >= len(partitionTable.Partitions) {
		return http.StatusBadRequest, []byte("Partition not found")
	}

	partition := partitionTable.Partitions[partitionIndex]

	image, err := hc.imageDir.FindImage(imageName)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	stream, err := image.OpenImage()
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}
	defer stream.Close()

	startOffset := int64(partitionTable.SectorSize) * int64(partition.Start)
	_, err = io.CopyN(io.Discard, stream, startOffset)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	w.WriteHeader(http.StatusOK)

	size := int64(partition.Size) * int64(partitionTable.SectorSize)
	compressor := compression.NewStreamCompressorN(w, stream, size, compressionBinary)
	if err = compressor.Run(); err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	return http.StatusOK, nil
}

// -----

func getPartitionTable(imageDir images.ImageDir, imageName string) (*disk.PartitionTable, error) {
	image, err := imageDir.FindImage(imageName)
	if err != nil {
		return nil, err
	}

	partitionTable, err := image.GetPartitionTable()
	if err != nil {
		return nil, err
	}

	return partitionTable, err
}
