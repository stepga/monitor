.PHONY: all daemon node

all: daemon node node_openbsd node_pi
daemon: bin/daemon
node: bin/node
node_openbsd: bin/node_openbsd
node_pi: bin/node_pi

GO_SOURCES := $(shell find . -name '*.go')
WEBUI_SOURCES := webui/assets/index.html $(wildcard webui/assets/static/*)

bin/daemon: cmd/daemon/main.go $(GO_SOURCES) $(WEBUI_SOURCES)
	go build -o $@ $<

bin/node: cmd/node/main.go $(GO_SOURCES)
	go build -o $@ $<

# If Go code requires CGO, the Go toolchain must invoke a C compiler and linker
# during the build. This make cross compiling to OpenBSD complex, because a
# OpenBSD gcc cross compiler (e.g. x86_64-unknown-openbsd) or zig is required.
# To avoid this, only use pure Go implementations/libs (via CGO_ENABLED=0).
bin/node_openbsd: cmd/node/main.go $(GO_SOURCES)
	env CGO_ENABLED=0 GOOS=openbsd go build -o $@ $<

bin/node_pi: cmd/node/main.go $(GO_SOURCES)
	GOOS=linux GOARCH=arm GOARM=7 go build -o $@ $<

clean:
	rm -rf bin
