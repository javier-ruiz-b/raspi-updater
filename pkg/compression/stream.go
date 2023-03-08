package compression

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
)

type CompressionStream struct {
	inStream     io.Reader
	sizeIn       int64
	stdoutStream io.ReadCloser
	command      string
	commandArgs  []string
	runningCmd   *exec.Cmd
	stderrBuffer bytes.Buffer
	finished     bool
}

var _ io.ReadSeekCloser = (*CompressionStream)(nil)

func (c *CompressionStream) Close() error {
	if c.runningCmd == nil {
		return fmt.Errorf("not running")
	}
	if c.runningCmd.ProcessState == nil || !c.runningCmd.ProcessState.Exited() {
		c.runningCmd.Process.Kill()
	}
	return nil
}

func (c *CompressionStream) Open() error {
	c.runningCmd = exec.Command(c.command, c.commandArgs...)

	if c.sizeIn == -1 {
		c.runningCmd.Stdin = c.inStream
	} else {
		c.runningCmd.Stdin = &io.LimitedReader{R: c.inStream, N: c.sizeIn}
	}

	var err error
	c.stdoutStream, err = c.runningCmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrBuffer := &bytes.Buffer{}
	c.runningCmd.Stderr = stderrBuffer

	c.finished = false

	err = c.runningCmd.Start()
	return err
}

func (c *CompressionStream) Read(p []byte) (n int, err error) {
	n, err = c.stdoutStream.Read(p)
	if err == io.EOF {
		c.finished = true
		cmdErr := c.runningCmd.Wait()
		if cmdErr != nil {
			err = fmt.Errorf("%v.  Stderr: %s", cmdErr, c.stderrBuffer.String())
		}
	}
	return n, err
}

func (c *CompressionStream) Seek(offset int64, whence int) (int64, error) {
	if whence != 1 {
		return -1, fmt.Errorf("seek allows only whence == 1")
	}
	return io.CopyN(io.Discard, c.stdoutStream, offset)
}
