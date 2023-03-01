package runner

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

func (r *FakeRunner) Run(file *os.File, args ...string) error {
	return r.RunPath(file.Name())
}

func (o *FakeRunner) RunPath(filePath string, args ...string) error {
	log.Print("Faking successful ", filePath, " execution")
	o.isRun = true
	return nil
}

func (r *FakeRunner) IsRun() bool {
	return r.isRun
}
