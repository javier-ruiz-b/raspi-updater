package compression

import "os/exec"

func CheckFile(compressor string, file string) error {
	cmd := exec.Command(compressor, "-t", file)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
