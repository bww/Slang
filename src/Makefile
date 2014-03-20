
export CC 		:= clang
export CXX 		:= clang++
export LIBDIR := $(shell pwd)/../lib
export GOPATH := $(GOPATH):$(LIBDIR)
export CGO_LDFLAGS := -lc -lc++ -L../dep/libsass -lsass -L../dep/jsmin -ljsmin

SOURCES=slang.go flags.go options.go compiler.go compiler_chain.go compiler_sass.go compiler_jsmin.go compiler_ejs.go compiler_literal.go server.go server_proxy.go
LIBRARY=$(LIBDIR)/src/ejs/scanner.go $(LIBDIR)/src/bww/errors/errors.go
BIN=../bin
SLANG=$(BIN)/slang

all: $(SLANG)

$(SLANG): $(SOURCES) $(LIBRARY)
	mkdir -p $(BIN)
	go build -o $(SLANG)

clean:
	rm -f $(SLANG)

