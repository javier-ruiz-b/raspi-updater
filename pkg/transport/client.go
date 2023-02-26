package transport

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/progress"
)

type Client interface {
	Close()
	Get(url string) (*http.Response, error)
	GetBytes(url string) ([]byte, error)
	GetString(url string) (string, error)
	DownloadFile(filepath string, url string, pr progress.Progress) error
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
	return c.tc.Get(url)
}

func (c *ClientStruct) GetBytes(url string) ([]byte, error) {
	body := &bytes.Buffer{}

	response, err := c.tc.Get(url)
	if response.StatusCode != http.StatusOK {
		return body.Bytes(), fmt.Errorf("error getting  %s, unexpected status code: %d", url, response.StatusCode)
	}
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

func (c *ClientStruct) DownloadFile(filepath string, url string, pr progress.Progress) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	response, err := c.tc.Get(url)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error getting file %s, unexpected status code: %d", url, response.StatusCode)
	}

	counter := progress.NewIoCounter(response.ContentLength, pr)
	_, err = io.Copy(out, io.TeeReader(response.Body, counter))
	return err
}
