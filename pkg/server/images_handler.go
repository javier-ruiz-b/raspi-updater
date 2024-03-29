package server

import (
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/compression"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/disk"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/images"
	"github.com/schollz/progressbar/v3"
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
	compressionExtension := mux.Vars(r)["compression"]
	partitionIndex, err := strconv.Atoi(mux.Vars(r)["partitionIndex"])
	if err != nil {
		return http.StatusBadRequest, []byte(err.Error())
	}
	tool, found := compression.AvailableToolByFileExtension()[compressionExtension]
	if !found {
		return http.StatusInternalServerError, []byte(fmt.Sprintf("compressor tool for " + compressionExtension + " is not available"))
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

	size := int64(partition.Size) * int64(partitionTable.SectorSize)
	bar := progressbar.DefaultBytes(size, "Sending "+imageName+" partition "+strconv.Itoa(partitionIndex))
	defer bar.Close()

	compressor := compression.NewStreamCompressorN(io.TeeReader(stream, bar), size, tool)
	if err = compressor.Open(); err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}
	defer compressor.Close()

	buffer := make([]byte, 1*1024*1024)
	if _, err = io.CopyBuffer(w, compressor, buffer); err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	return http.StatusOK, nil
}

func (hc *HandlerConfig) imageBackupExists(w http.ResponseWriter, r *http.Request) (int, []byte) {
	imageName := mux.Vars(r)["id"]
	version := mux.Vars(r)["version"]

	backupExists := hc.imageDir.BackupExists(imageName, version)

	err := gob.NewEncoder(w).Encode(backupExists)
	if err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	return http.StatusOK, nil
}

func (hc *HandlerConfig) imageBackup(w http.ResponseWriter, r *http.Request) (int, []byte) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, []byte("Method not allowed")
	}

	imageName := mux.Vars(r)["id"]
	compressionExtension := mux.Vars(r)["compression"]
	if imageName == "" || compressionExtension == "" {
		return http.StatusBadRequest, []byte("Missing parameters")
	}
	tool, found := compression.AvailableToolByFileExtension()[compressionExtension]
	if !found {
		return http.StatusInternalServerError, []byte(fmt.Sprintf("compressor tool for " + compressionExtension + " is not available"))
	}
	version := mux.Vars(r)["version"]

	outFile, err := hc.imageDir.CreateBackup(imageName, version, tool)
	if err != nil {
		return http.StatusInternalServerError, []byte(fmt.Sprintf("Failed to create backup: %v", err))
	}
	defer outFile.Close()

	// Test compression on the fly
	streamTestReader, streamTestWriter := io.Pipe()
	tester := compression.NewStreamTester(streamTestReader, tool)
	if err := tester.Open(); err != nil {
		return http.StatusInternalServerError, []byte(fmt.Sprintf("Failed to open stream tester with %s: %s", tool, err.Error()))
	}
	defer tester.Close()

	bar := progressbar.DefaultBytes(-1, "Saving backup "+imageName)
	defer bar.Close()

	// Copy the file data to the output file
	buffer := make([]byte, 1*1024*1024)
	if _, err = io.CopyBuffer(io.MultiWriter(outFile, streamTestWriter, bar), r.Body, buffer); err != nil {
		return http.StatusInternalServerError, []byte(fmt.Sprintf("Failed to create backup %s", outFile.Name()))
	}

	if err = streamTestWriter.Close(); err != nil {
		return http.StatusInternalServerError, []byte("Can't close compression tester stream")
	}

	if _, err = tester.Read(make([]byte, 1)); err != io.EOF && err != nil {
		return http.StatusInternalServerError, []byte(fmt.Sprintf("Failed testing compressed stream: %s", err.Error()))
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
