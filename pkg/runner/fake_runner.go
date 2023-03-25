package runner

import (
	"log"
	"os"
	"strings"
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
	log.Print("Faking successful ", filePath, " ", strings.Join(args, " "), " execution")
	o.isRun = true
	return nil
}

func (r *FakeRunner) IsRun() bool {
	return r.isRun
}
