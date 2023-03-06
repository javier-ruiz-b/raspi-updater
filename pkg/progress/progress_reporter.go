package progress

import (
	"fmt"
	"io"
	"os"
	"time"
)

func NewMainProgressReporter() Progress {
	result := NewProgressReporter(nil, 100)
	return result
}

type ProgressReporter struct {
	parent      Progress
	Stdout      io.Writer
	description string
	percent     int
	minPercent  int
	maxPercent  int
	lastUpdate  int64
}

func NewProgressReporter(parentReporter Progress, maxPercent int) *ProgressReporter {
	minPercent := 0
	if parentReporter != nil {
		minPercent = parentReporter.Percent()
	}

	return &ProgressReporter{
		Stdout:      os.Stdout,
		description: "",
		parent:      parentReporter,
		minPercent:  minPercent,
		maxPercent:  maxPercent,
	}
}

//
// Main		 - Partitioning		- mbr
// 0 - 100	 - 10 - 20

// Initializing -

func (pr *ProgressReporter) SetDescription(description string, percent int) {
	pr.UpdateDescription(description, percent)
	fmt.Fprintf(pr.Stdout, "\n")
}

func (pr *ProgressReporter) UpdateDescription(description string, percent int) {
	pr.description = description
	pr.SetPercent(percent)
	pr.updateStdout()
}

func (pr *ProgressReporter) SetPercent(percent int) {
	maxRange := pr.maxPercent - pr.minPercent
	pr.percent = pr.minPercent + (percent * maxRange / 100)
	pr.updateStdout()
}

func (pr *ProgressReporter) Percent() int {
	return pr.percent
}

func (pr *ProgressReporter) Description() string {
	if pr.parent != nil {
		return pr.parent.Description() + ": " + pr.description
	}
	return pr.description
}

func (pr *ProgressReporter) Printf(format string, a ...any) {
	fmt.Fprintf(pr.Stdout, "\r"+format+"        \n", a...)
	pr.updateStdout()
}

func (pr *ProgressReporter) updateStdout() {
	nowUnix := time.Now().Unix()
	if pr.lastUpdate == nowUnix {
		return
	}
	fmt.Fprintf(pr.Stdout, "\r [%3d%% ] %s        ", pr.Percent(), pr.Description())
	pr.lastUpdate = nowUnix
}
