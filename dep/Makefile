
all: golib
	cd libsass && make
	cd jsmin && make

golib:
	go get github.com/BurntSushi/toml

clean:
	cd libsass && make clean
	cd jsmin && make clean

