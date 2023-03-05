package compression

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

func commandPipe(command string, args ...string) (io.WriteCloser, io.ReadCloser, chan error) {
	cmd := exec.Command(command, args...)
	errChannel := make(chan error, 1)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		errChannel <- err
		return nil, nil, errChannel
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errChannel <- err
		return nil, nil, errChannel
	}

	go func() {
		stderrBuffer := &bytes.Buffer{}
		cmd.Stderr = stderrBuffer
		err := cmd.Run()
		if err != nil {
			errChannel <- fmt.Errorf("%v.  Stderr: %s", err, stderrBuffer.String())
		}
		errChannel <- nil
	}()
	return stdin, stdout, errChannel
}
