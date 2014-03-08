
export CC := clang
export CGO_LDFLAGS := -lc -lc++ -L../dep/libsass -lsass -L../dep/jsmin -ljsmin

SOURCES=webasm.go flags.go compiler.go compiler_sass.go compiler_jsmin.go server.go server_proxy.go

all: webasm

webasm: $(SOURCES)
	go build -o webasm

clean:
	rm -f webasm

