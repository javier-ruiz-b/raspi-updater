package progress

import "fmt"

func NewMainProgressReporter() Progress {
	result := NewProgressReporter(nil, 100)
	result.SetDescription("Initializing...", 0)
	return result
}

type ProgressReporter struct {
	parent      Progress
	description string
	percent     int
	minPercent  int
	maxPercent  int
}

func NewProgressReporter(parentReporter Progress, maxPercent int) *ProgressReporter {
	minPercent := 0
	if parentReporter != nil {
		minPercent = parentReporter.Percent()
	}

	return &ProgressReporter{
		parent:     parentReporter,
		minPercent: minPercent,
		maxPercent: maxPercent,
	}
}

//
// Main		 - Partitioning		- mbr
// 0 - 100	 - 10 - 20

// Initializing -

func (pr *ProgressReporter) SetDescription(description string, percent int) {
	fmt.Println()
	pr.description = description
	pr.SetPercent(percent)
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
	fmt.Printf("\\33[2K\r"+format+"\n", a...)
	pr.updateStdout()
}

func (pr *ProgressReporter) updateStdout() {
	fmt.Printf("\\33[2K\r [ %3d%% ] %s", pr.Percent(), pr.Description())
}
