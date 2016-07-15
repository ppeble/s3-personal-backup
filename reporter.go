package backup

type Reporter interface {
	Run()
	Print()
}
