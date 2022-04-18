package server

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func (h *HandlerConfig) updateHandler(w http.ResponseWriter, r *http.Request) {
	filepath := mux.Vars(r)["filename"]
	file, err := os.Open(h.binariesDir + "/" + filepath)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	}

	io.Copy(w, file)
}
