package selfupdater

import "os"

type Runner interface {
	Run(file *os.File) error
}
