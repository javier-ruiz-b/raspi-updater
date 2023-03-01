package selfupdater

import (
	"os"
	"runtime"
	"strings"

	"github.com/javier-ruiz-b/raspi-image-updater/pkg/progress"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/runner"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/server"
	"github.com/javier-ruiz-b/raspi-image-updater/pkg/transport"
)

type SelfUpdater struct {
	client transport.Client
	runner runner.Runner
}

func NewSelfUpdater(client transport.Client, runner runner.Runner) *SelfUpdater {
	return &SelfUpdater{
		client: client,
		runner: runner,
	}
}

func (u *SelfUpdater) DownloadAndRunUpdate(progress progress.Progress) error {
	binaryFile, err := u.downloadBinary(progress) // TODO: add progress functionality
	if err != nil {
		return err
	}

	return u.runner.Run(binaryFile)
}

func (u *SelfUpdater) IsUpdateAvailable(clientVersion string) (bool, error) {
	serverVersion, err := u.client.GetString(server.API_VERSION)
	if err != nil {
		return false, err
	}

	return serverVersion != clientVersion, nil
}

func (u *SelfUpdater) downloadBinary(progress progress.Progress) (*os.File, error) {
	tempFile, err := os.CreateTemp(os.TempDir(), "updater")
	if err != nil {
		return nil, err
	}
	tempFile.Close()

	filename := runtime.GOOS + "-" + runtime.GOARCH
	url := strings.Replace(server.API_UPDATE, "{filename}", filename, 1)

	progress.SetPercent(0)
	err = u.client.DownloadFile(tempFile.Name(), url, progress)
	tempFile.Chmod(0777)
	return tempFile, err
}
