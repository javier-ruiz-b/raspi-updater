package transport

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"mime/multipart"
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
	UploadStream(url string, stream io.Reader) error
}

type ClientStruct struct {
	tc *QuicClient
}

func newClient(client *QuicClient) Client {
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

func (c *ClientStruct) UploadStream(url string, stream io.Reader) error {
	body, writer := io.Pipe()
	request, err := c.tc.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}

	mwriter := multipart.NewWriter(writer)

	request.Header.Add("Content-Type", mwriter.FormDataContentType())

	errChan := make(chan error, 1)
	go func() {
		defer mwriter.Close()
		// Create a new form file field
		fileField, err := mwriter.CreateFormFile("file", "file")
		if err != nil {
			errChan <- err
			return
		}

		if _, err = io.Copy(fileField, stream); err != nil {
			errChan <- err
			return
		}

		if err := mwriter.Close(); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	// Send the HTTP request and get the response
	response, err := c.tc.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Check the response status code for success
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error uploading stream %s, unexpected status code: %d", url, response.StatusCode)
	}
	if err := <-errChan; err != nil {
		return err
	}

	return err
}
