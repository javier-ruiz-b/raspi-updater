package transport

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/schollz/progressbar/v3"
)

type Client interface {
	Close()
	Get(url string) (*http.Response, error)
	GetBytes(url string) ([]byte, error)
	GetString(url string) (string, error)
	GetObject(url string, object any) error
	DownloadFile(filepath string, url string) error
	GetDownloadStream(url string) (io.ReadCloser, int64, error)
}

type ClientStruct struct {
	tc transportClient
}

func newClient(client transportClient) Client {
	return &ClientStruct{tc: client}
}

func (c *ClientStruct) Close() {
	c.tc.Close()
}

func (c *ClientStruct) Get(url string) (*http.Response, error) {
	response, err := c.tc.Get(url)
	if response.StatusCode != http.StatusOK {
		return response, fmt.Errorf("error getting  %s, unexpected status code: %d", url, response.StatusCode)
	}
	return response, err
}

func (c *ClientStruct) GetObject(url string, object any) error {
	response, err := c.tc.Get(url)
	if err != nil {
		return err
	}

	return gob.NewDecoder(response.Body).Decode(object)
}

func (c *ClientStruct) GetBytes(url string) ([]byte, error) {
	body := &bytes.Buffer{}

	response, err := c.tc.Get(url)
	if err != nil {
		return body.Bytes(), err
	}

	_, err = io.Copy(body, response.Body)
	return body.Bytes(), err
}

func (c *ClientStruct) GetString(url string) (string, error) {
	bytes, err := c.GetBytes(url)
	return string(bytes), err
}

func (c *ClientStruct) DownloadFile(filepath string, url string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	responseStream, contentLength, err := c.GetDownloadStream(url)
	if err != nil {
		return err
	}
	defer responseStream.Close()

	bar := progressbar.DefaultBytes(
		contentLength,
		"Downloading",
	)
	_, err = io.Copy(out, io.TeeReader(responseStream, bar))
	return err
}

func (c *ClientStruct) GetDownloadStream(url string) (io.ReadCloser, int64, error) {
	response, err := c.tc.Get(url)
	if err != nil {
		return nil, -1, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, -1, fmt.Errorf("error getting stream %s, unexpected status code: %d", url, response.StatusCode)
	}

	return response.Body, response.ContentLength, nil
}
