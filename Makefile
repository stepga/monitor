.PHONY: all cross_build

all: daemon node cross_build

daemon:
	go build -o bin/daemon cmd/daemon/main.go

node:
	go build -o bin/node cmd/node/main.go

cross_build:
	env GOOS=linux   GOARCH=amd64 go build -o bin/daemon_linux_amd64      cmd/daemon/main.go
	env GOOS=linux   GOARCH=amd64 go build -o bin/node_linux_amd64        cmd/node/main.go
	env GOOS=openbsd GOARCH=amd64 go build -o bin/node_openbsd_obsd_amd64 cmd/node/main.go
