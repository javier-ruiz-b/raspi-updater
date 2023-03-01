package progress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPercent(t *testing.T) {
	tested := NewMainProgressReporter()
	tested.SetPercent(49)

	assert.Equal(t, 49, tested.Percent())
}

func TestSubProgressPercent(t *testing.T) {
	parent := NewMainProgressReporter()
	parent.SetPercent(50)

	testedSub := NewProgressReporter(parent, 100)
	testedSub.SetPercent(50)

	assert.Equal(t, 75, testedSub.Percent())
}

func TestDescription(t *testing.T) {
	tested := NewMainProgressReporter()
	tested.SetDescription("Init", 0)

	assert.Equal(t, "Init", tested.Description())
}

func TestSubProgresstDescription(t *testing.T) {
	parent := NewMainProgressReporter()
	parent.SetDescription("Foo", 0)

	tested := NewProgressReporter(parent, 100)
	tested.SetDescription("Bar", 0)

	assert.Equal(t, "Foo: Bar", tested.Description())
}
