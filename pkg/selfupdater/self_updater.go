package selfupdater

import (
	"os"
	"runtime"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/transport"
)

type SelfUpdater struct {
	client transport.Client
	runner Runner
}

func NewSelfUpdater(client transport.Client, runner Runner) *SelfUpdater {
	return &SelfUpdater{
		client: client,
		runner: runner,
	}
}

func (u *SelfUpdater) DownloadAndRunUpdate() error {
	binaryFile, err := u.downloadBinary()
	if err != nil {
		return err
	}

	return u.runner.Run(binaryFile)
}

func (u *SelfUpdater) IsUpdateAvailable(clientVersion string) (bool, error) {
	serverVersion, err := u.client.GetString("/version")
	if err != nil {
		return false, err
	}

	return serverVersion != clientVersion, nil
}

func (u *SelfUpdater) downloadBinary() (*os.File, error) {
	tempFile, err := os.CreateTemp(os.TempDir(), "updater")
	if err != nil {
		return nil, err
	}
	tempFile.Close()

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	url := "/update/" + goos + "-" + goarch

	err = u.client.DownloadFile(tempFile.Name(), url)
	return tempFile, err
}
