
CC = cc
CFLAGS = -Wall -O2 -fPIC -g
LDFLAGS = -fPIC

PREFIX    = /usr/local
LIBDIR    = $(PREFIX)/lib

SOURCES = slang_jsmin.c
OBJECTS = $(SOURCES:.c=.o)

all: static

static: libjsmin.a
shared: libjsmin.so

libjsmin.a: $(OBJECTS)
	ar rvs $@ $(OBJECTS)

libjsmin.so: $(OBJECTS)
	$(CC) -shared $(LDFLAGS) -o $@ $(OBJECTS)

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

%: %.o libjsmin.a
	$(CC) $(CFLAGS) -o $@ $+ $(LDFLAGS)

clean:
	rm -f $(OBJECTS) *.a *.so

