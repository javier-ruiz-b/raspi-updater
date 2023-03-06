package progress

type Progress interface {
	SetPercent(int)
	UpdateDescription(string, int)
	SetDescription(string, int)
	Percent() int
	Description() string
	Printf(format string, a ...any)
}
