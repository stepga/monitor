package subsystems

type Subsystem interface {
	Init() error
}

type Report interface {
	Report() string
}

type Summary interface {
	Summary() string
}
