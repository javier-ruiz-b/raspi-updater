package progress

import (
	"fmt"
	"io"

	humanize "github.com/dustin/go-humanize"
)

type IoCounter struct {
	Transferred   uint64
	contentLength int64
	pr            Progress
}

var _ io.WriteCloser = (*IoCounter)(nil)

func NewIoCounter(contentLength int64, pr Progress) *IoCounter {
	pr.SetPercent(0)
	return &IoCounter{
		Transferred:   0,
		contentLength: contentLength,
		pr:            pr,
	}
}

func (ic *IoCounter) Write(p []byte) (int, error) {
	n := len(p)
	ic.Transferred += uint64(n)

	var description string
	progressPercent := 0
	if ic.contentLength >= 0 {
		description = fmt.Sprintf("Transferred %s of %s", humanize.Bytes(ic.Transferred), humanize.Bytes(uint64(ic.contentLength)))
		progressPercent = int(int64(ic.Transferred) * 100 / ic.contentLength)
	} else {
		description = fmt.Sprintf("Transferred %s", humanize.Bytes(ic.Transferred))
	}
	ic.pr.UpdateDescription(description, progressPercent)

	return n, nil
}

func (ic *IoCounter) Close() error {
	ic.pr.SetDescription("Finished", 100)
	return nil
}
