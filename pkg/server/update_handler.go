package server

import (
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func (hc *HandlerConfig) updateHandler(w http.ResponseWriter, r *http.Request) (int, []byte) {
	filepath := mux.Vars(r)["filename"]
	file, err := os.Open(hc.binariesDir + "/" + filepath)
	if err != nil {
		return http.StatusNotFound, []byte(err.Error())
	}

	buffer := make([]byte, 1*1024*1024) // 1 MB
	if _, err = io.CopyBuffer(w, file, buffer); err != nil {
		return http.StatusInternalServerError, []byte(err.Error())
	}

	return http.StatusOK, nil
}
