package reporter

type Report interface {
	Report() string
}

type Reporter interface {
	Init()
}
