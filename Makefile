.PHONY: all mkdirs

all: daemon

mkdirs:
	@mkdir -p bin

daemon: mkdirs
	go build -o bin/daemon cmd/daemon/main.go

node: mkdirs
	go build -o bin/node cmd/node/main.go
