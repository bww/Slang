
export CC := clang
export CGO_LDFLAGS := -L../dep/libsass -lc -lc++ -lsass

SOURCES=webasm.go compiler.go compiler_sass.go

all: webasm

webasm: $(SOURCES)
	go build -o webasm

clean:
	rm -f webasm

