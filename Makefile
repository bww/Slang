
export BUILD 				:= $(shell pwd)
export GOPATH 			:= $(GOPATH):$(BUILD)
export CGO_LDFLAGS 	:= -lc -lc++ -L$(BUILD)/dep/libsass -lsass -L$(BUILD)/dep/jsmin -ljsmin

BIN=bin
SLANG=$(BIN)/slang

SOURCES=\
	src/main/slang.go \
	src/main/flags.go \
	src/main/options.go \
	src/main/compiler.go \
	src/main/compiler_chain.go \
	src/main/compiler_sass.go \
	src/main/compiler_jsmin.go \
	src/main/compiler_ejs.go \
	src/main/compiler_literal.go \
	src/main/server.go \
	src/main/server_proxy.go \
	src/ejs/scanner.go \
	src/bww/errors/errors.go

all: deps $(SLANG)

deps:
	cd dep && make
	go get github.com/BurntSushi/toml
	go get bitbucket.org/kardianos/osext

$(SLANG): $(SOURCES)
	mkdir -p $(BIN)
	go build -o $(SLANG) ./src/main

clean:
	rm -f $(SLANG)

cleanall: clean
	cd dep && make clean

