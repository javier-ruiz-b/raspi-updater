package runner

import "os"

type Runner interface {
	Run(file *os.File, args ...string) error
	RunPath(filePath string, args ...string) error
}
