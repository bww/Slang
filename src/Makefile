
export CC 		:= clang
export LIBDIR := $(shell pwd)/../lib
export GOPATH := $(GOPATH):$(LIBDIR)
export CGO_LDFLAGS := -lc -lc++ -L../dep/libsass -lsass -L../dep/jsmin -ljsmin

SOURCES=webasm.go flags.go compiler.go compiler_sass.go compiler_jsmin.go compiler_ejs.go compiler_literal.go server.go server_proxy.go
LIBRARY=$(LIBDIR)/src/ejs/scanner.go

all: webasm

webasm: $(SOURCES) $(LIBRARY)
	go build -o webasm

clean:
	rm -f webasm

