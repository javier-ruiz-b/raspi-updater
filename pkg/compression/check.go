package compression

import "os/exec"

func CheckFile(compressor string, file string) error {
	return exec.Command(compressor, "-t", file).Run()
}
