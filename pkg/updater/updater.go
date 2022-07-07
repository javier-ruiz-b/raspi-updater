package updater

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/network"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/transport"
)

type Updater struct {
	client transport.Client
	runner Runner
}

func NewUpdater(client transport.Client, runner Runner) *Updater {
	return &Updater{
		client: client,
		runner: runner,
	}
}

func (u *Updater) DownloadAndRunUpdate() error {
	binaryFile, err := u.downloadBinary()
	if err != nil {
		return err
	}

	return u.runner.Run(binaryFile)
}

func (u *Updater) IsUpdateAvailable(clientVersion string) (bool, error) {
	response, err := u.client.Get("/version")
	if err != nil {
		return false, err
	}

	serverVersion, err := transport.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	return string(serverVersion) != clientVersion, nil
}

func (u *Updater) downloadBinary() (*os.File, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	url := "/update/" + goos + "-" + goarch

	response, err := u.client.Get(url)
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
