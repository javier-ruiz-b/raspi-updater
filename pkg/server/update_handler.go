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

	io.Copy(w, file)
	return http.StatusOK, nil
}
