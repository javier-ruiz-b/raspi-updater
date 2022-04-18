package client

import (
	"bytes"
	"errors"
	"io"
	"log"
)

func update(qc *QuicClient, clientVersion string) error {
	response, err := qc.Get("/version")
	if err != nil {
		log.Print(err)
		return err
	}
	serverVersion, err := readAll(response.Body)
	if err != nil {
		return err
	}
	if string(serverVersion) == clientVersion {
		return nil
	}

	return errors.New("unimplemented")

}

func readAll(src io.Reader) ([]byte, error) {
	body := &bytes.Buffer{}
	_, err := io.Copy(body, src)
	return body.Bytes(), err
}
