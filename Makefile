
export BUILD 				:= $(shell pwd)
export GOPATH 			:= $(GOPATH):$(BUILD)
export CGO_LDFLAGS 	:= -lc -lc++ -L$(BUILD)/dep/libsass -lsass -L$(BUILD)/dep/jsmin -ljsmin

BIN=bin
SLANG=$(BIN)/slang

SOURCES=\
	src/main/*.go \
	src/ejs/*.go \
	src/bww/errors/*.go

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

