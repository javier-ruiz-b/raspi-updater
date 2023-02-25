package selfupdater

import (
	"log"
	"os"
)

type FakeRunner struct {
	isRun bool
}

func NewFakeRunner() *FakeRunner {
	return &FakeRunner{isRun: false}
}

func (r *FakeRunner) Run(file *os.File) error {
	log.Print("Faking successful ", file.Name(), " execution")
	r.isRun = true
	return nil
}

func (r *FakeRunner) IsRun() bool {
	return r.isRun
}
