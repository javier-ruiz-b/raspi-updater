package selfupdater

import (
	"fmt"
	"os"
	"runtime"
)

type OsRunner struct{}

func (*OsRunner) Run(file *os.File) error {
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
