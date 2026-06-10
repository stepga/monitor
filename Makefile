.PHONY: all

all: daemon node node_openbsd

daemon: cmd/daemon/main.go
	go build -o bin/$@ $<

node: cmd/node/main.go
	go build -o bin/$@ $<

# If Go code requires CGO, the Go toolchain must invoke a C compiler and linker
# during the build. This make cross compiling to OpenBSD complex, because a
# OpenBSD gcc cross compiler (e.g. x86_64-unknown-openbsd) or zig is required.
# To avoid this, only use pure Go implementations/libs (via CGO_ENABLED=0).
node_openbsd: cmd/node/main.go
	env CGO_ENABLED=0 GOOS=openbsd go build -o bin/$@ $<

clean:
	rm -rf bin
