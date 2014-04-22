
PREFIX=/opt/slang

.PHONY: all static shared install clean

all: static

static:
	cd libsass && make static
	cd jsmin && make static

shared:
	cd libsass && make shared
	cd jsmin && make shared

install: shared
	install -D libsass/libsass.so $(PREFIX)/lib/libsass.so
	install -D jsmin/libjsmin.so $(PREFIX)/lib/libjsmin.so

clean:
	cd libsass && make clean
	cd jsmin && make clean

