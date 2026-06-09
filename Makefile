.PHONY: all

all: daemon node

daemon:
	go build -o bin/daemon ./cmd/daemon/

node:
	go build -o bin/node ./cmd/node/
