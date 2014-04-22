
OS=$(shell uname -s)

ifeq ($(OS),Linux)
TARGET=shared
else
TARGET=static
endif

.PHONY: all build install clean

all: build

build:
	cd libsass && make $(TARGET)
	cd jsmin && make

install: build
	install -D libsass/libsass.so /opt/libsass/lib/libsass.so

clean:
	cd libsass && make clean
	cd jsmin && make clean

