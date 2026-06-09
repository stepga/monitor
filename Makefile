.PHONY: all

all: daemon node

daemon:
	go build -o bin/daemon ./cmd/daemon/
	env GOOS=linux GOARCH=amd64 go build -o bin/daemon cmd/daemon/main.go

node:
	env GOOS=linux GOARCH=amd64 go build -o bin/node_linux cmd/node/main.go
	env GOOS=openbsd GOARCH=amd64 go build -o bin/node_openbsd cmd/node/main.go
