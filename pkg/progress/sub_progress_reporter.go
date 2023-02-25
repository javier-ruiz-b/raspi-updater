package progress

import "fmt"

type SubProgressReporter struct {
	parent      Progress
	description string
	percent     int
	minPercent  int
	maxPercent  int
}

func NewSubProgressReporter(parentReporter Progress, maxPercent int) *SubProgressReporter {
	minPercent := 0
	if parentReporter != nil {
		minPercent = parentReporter.Percent()
	}

	return &SubProgressReporter{
		parent:     parentReporter,
		minPercent: minPercent,
		maxPercent: maxPercent,
	}
}

//
// Main		 - Partitioning		- mbr
// 0 - 100	 - 10 - 20

// Initializing -

func (pr *SubProgressReporter) SetDescription(description string, percent int) {
	pr.description = description
	pr.SetPercent(percent)
}

func (pr *SubProgressReporter) SetPercent(percent int) {
	maxRange := pr.maxPercent - pr.minPercent
	pr.percent = pr.minPercent + (percent * maxRange / 100)
	pr.updateStdout()
}

func (pr *SubProgressReporter) Percent() int {
	return pr.percent
}

func (pr *SubProgressReporter) Description() string {
	if pr.parent != nil {
		return pr.parent.Description() + ": " + pr.description
	}
	return pr.description
}

func (pr *SubProgressReporter) updateStdout() {
	fmt.Printf("\r [ %3d%% ] %s", pr.Percent(), pr.Description())
}
