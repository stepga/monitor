.PHONY: all daemon node

all: daemon node node_openbsd
daemon: bin/daemon
node: bin/node
node_openbsd: bin/node_openbsd

NODE_DEPS :=$(shell find ./node -name '*.go')
DAEMON_DEPS :=$(shell find . -name '*.go' -not -path './node/*' -not -path './cmd/node/*')

bin/daemon: cmd/daemon/main.go $(DAEMON_DEPS)
	go build -o $@ $<

bin/node: cmd/node/main.go $(NODE_DEPS)
	go build -o $@ $<

# If Go code requires CGO, the Go toolchain must invoke a C compiler and linker
# during the build. This make cross compiling to OpenBSD complex, because a
# OpenBSD gcc cross compiler (e.g. x86_64-unknown-openbsd) or zig is required.
# To avoid this, only use pure Go implementations/libs (via CGO_ENABLED=0).
bin/node_openbsd: cmd/node/main.go $(NODE_DEPS)
	env CGO_ENABLED=0 GOOS=openbsd go build -o $@ $<

clean:
	rm -rf bin
