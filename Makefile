
export BUILD 	:= $(PWD)
export GOPATH := $(GOPATH):$(BUILD)

OS=$(shell uname -s)

ifeq ($(OS),Linux)
export CGO_LDFLAGS 	:= -L/opt/slang/lib -lsass -ljsmin
DEPS_TARGET=install
else
export CGO_LDFLAGS 	:= -lc -lc++ -L$(BUILD)/dep/libsass -lsass -L$(BUILD)/dep/jsmin -ljsmin
DEPS_TARGET=static
endif

BIN=bin
SLANG=$(BIN)/slang
PREFIX=/opt/slang

SOURCES=\
	src/main/*.go \
	src/ejs/*.go \
	src/bww/errors/*.go

.PHONY all deps install clean cleanall

all: deps $(SLANG)

deps:
	cd dep && make PREFIX=$(PREFIX) $(DEPS_TARGET)
	go get github.com/BurntSushi/toml
	go get bitbucket.org/kardianos/osext

$(SLANG): $(SOURCES)
	mkdir -p $(BIN)
	go build -o $(SLANG) ./src/main

install: $(SLANG)
	install -D $(SLANG) $(PREFIX)/bin/slang

clean:
	rm -f $(SLANG)

cleanall: clean
	cd dep && make clean

