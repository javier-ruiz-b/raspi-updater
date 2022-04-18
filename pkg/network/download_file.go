package network

import (
	"io"
	"net/http"
	"os"
)

func DownloadFile(filepath string, response *http.Response) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, response.Body)
	return err
}
