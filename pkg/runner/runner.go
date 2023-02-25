package runner

import "os"

type Runner interface {
	Run(file *os.File) error
}
