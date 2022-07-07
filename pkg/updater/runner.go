package updater

import "os"

type Runner interface {
	Run(file *os.File) error
}
