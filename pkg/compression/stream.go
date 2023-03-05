package compression

import (
	"fmt"
	"io"
)

type CompressionStream struct {
	inStream    io.Reader
	sizeIn      int64
	outStream   io.Writer
	sizeOut     int64
	command     string
	commandArgs []string
}

func (c *CompressionStream) Run() error {
	stdin, stdout, commandErrChannel := commandPipe(c.command, c.commandArgs...)
	resultChannel := make(chan error)
	go func() {
		var copied int64 = -1
		var err error

		if c.sizeIn == -1 {
			_, err = io.Copy(stdin, c.inStream)
		} else {
			copied, err = io.CopyN(stdin, c.inStream, c.sizeIn)
		}

		stdin.Close()
		if c.sizeIn != copied {
			resultChannel <- fmt.Errorf("did not copy all contents. Expected: %d, copied: %d", c.sizeIn, copied)
			return
		}
		resultChannel <- err
	}()

	var copied int64 = -1
	var err error
	if c.sizeOut == -1 {
		_, err = io.Copy(c.outStream, stdout)
	} else {
		copied, err = io.CopyN(c.outStream, stdout, c.sizeOut)
	}

	commandErr := <-commandErrChannel
	if commandErr != nil {
		return commandErr
	}

	compressErr := <-resultChannel
	if compressErr != nil {
		return compressErr
	}

	if c.sizeOut != copied {
		return fmt.Errorf("did not copy all contents. Expected: %d, copied: %d", c.sizeOut, copied)
	}

	return err
}
