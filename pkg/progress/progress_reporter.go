package progress

func NewProgressReporter() Progress {
	result := NewSubProgressReporter(nil, 100)
	result.SetDescription("Initializing...", 0)
	return result
}
