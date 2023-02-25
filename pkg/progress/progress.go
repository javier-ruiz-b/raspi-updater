package progress

type Progress interface {
	SetPercent(int)
	SetDescription(string, int)
	Percent() int
	Description() string
}
