package runner

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type OsRunner struct{}

func (o *OsRunner) RunPath(filePath string, args ...string) error {
	cmd := exec.Command(filePath, args...)
	err := cmd.Run()
	if err != nil {
		errExit := err.(*exec.ExitError)
		return fmt.Errorf("process %s ended with exit code %d. Output: %s", filePath, errExit.ExitCode(), errExit.Stderr)
	}
	return err
}

func (o *OsRunner) Run(file *os.File, args ...string) error {
	path := file.Name()
	if runtime.GOOS == "windows" {
		err := os.Rename(path, path+".exe")
		if err != nil {
			return err
		}
		path += ".exe"
	}
	defer os.Remove(path)

	return o.RunPath(path)
}
