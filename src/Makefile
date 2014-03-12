
export CC 		:= clang
export LIBDIR := $(shell pwd)/../lib
export GOPATH := $(GOPATH):$(LIBDIR)
export CGO_LDFLAGS := -lc -lc++ -L../dep/libsass -lsass -L../dep/jsmin -ljsmin

SOURCES=webasm.go flags.go options.go compiler.go compiler_chain.go compiler_sass.go compiler_jsmin.go compiler_ejs.go compiler_literal.go server.go server_proxy.go
LIBRARY=$(LIBDIR)/src/ejs/scanner.go
WEBASM=../webasm

all: $(WEBASM)

$(WEBASM): $(SOURCES) $(LIBRARY)
	go build -o $(WEBASM)

clean:
	rm -f $(WEBASM)

