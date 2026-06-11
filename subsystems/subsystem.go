package subsystems

type Subsystem interface {
	Init()
}

type Report interface {
	Report() string
}
