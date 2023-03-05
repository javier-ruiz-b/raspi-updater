package progress

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var stdoutBuf bytes.Buffer

func TestPercent(t *testing.T) {
	tested := newProgressTested()
	tested.SetPercent(49)

	assert.Equal(t, 49, tested.Percent())
}

func TestSubProgressPercent(t *testing.T) {
	parent := newProgressTested()
	parent.SetPercent(50)

	testedSub := NewProgressReporter(parent, 100)
	testedSub.SetPercent(50)

	assert.Equal(t, 75, testedSub.Percent())
}

func TestDescription(t *testing.T) {
	tested := newProgressTested()
	tested.SetDescription("Init", 0)

	assert.Equal(t, "Init", tested.Description())
	assert.Equal(t, "\n\\33[2K\r [  0% ] Init", stdoutBuf.String())
}

func TestSubProgresstDescription(t *testing.T) {
	parent := newProgressTested()
	parent.SetDescription("Foo", 0)

	tested := NewProgressReporter(parent, 100)
	tested.SetDescription("Bar", 0)

	assert.Equal(t, "Foo: Bar", tested.Description())
}

func newProgressTested() Progress {
	tested := NewMainProgressReporter().(*ProgressReporter)
	stdoutBuf = bytes.Buffer{}
	tested.Stdout = &stdoutBuf
	return tested
}
