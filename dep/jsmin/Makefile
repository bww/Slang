
CC = cc
CCFLAGS = -Wall -O2 -lc
LDFLAGS =

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

%.o: %.cpp
	$(CC) $(CCFLAGS) -c -o $@ $<

%: %.o libjsmin.a
	$(CC) $(CCFLAGS) -o $@ $+ $(LDFLAGS)

clean:
	rm -f $(OBJECTS) *.a *.so

