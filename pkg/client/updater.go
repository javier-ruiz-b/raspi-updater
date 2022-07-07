package client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/network"
)

func update(qc *QuicClient, clientVersion string) error {
	response, err := qc.Get("/version")
	if err != nil {
		return err
	}
	serverVersion, err := readAll(response.Body)
	if err != nil {
		return err
	}
	if string(serverVersion) == clientVersion {
		return nil
	}

	updateFile, err := downloadUpdate(qc)
	if err != nil {
		return err
	}

	return runUpdate(updateFile)
}

func runUpdate(file *os.File) error {
	file.Chmod(0777)

	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}

	path := file.Name()
	if runtime.GOOS == "windows" {
		err := os.Rename(path, path+".exe")
		if err != nil {
			return err
		}
		path += ".exe"
	}
	defer os.Remove(path)

	process, err := os.StartProcess(path, os.Args, &procAttr)
	if err != nil {
		return err
	}

	state, err := process.Wait()
	if state.ExitCode() != 0 {
		return fmt.Errorf("process %s ended with exit code %d", file.Name(), state.ExitCode())
	}

	return err
}

func downloadUpdate(qc *QuicClient) (*os.File, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	url := "/update/" + goos + "-" + goarch

	response, err := qc.Get(url)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting %s, unexpected status code: %d", url, response.StatusCode)
	}

	tempFile, err := ioutil.TempFile(os.TempDir(), "updater")
	if err != nil {
		return nil, err
	}
	tempFile.Close()

	err = network.DownloadFile(tempFile.Name(), response)
	return tempFile, err
}

func readAll(src io.Reader) ([]byte, error) {
	body := &bytes.Buffer{}
	_, err := io.Copy(body, src)
	return body.Bytes(), err
}
