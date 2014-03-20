
all: golib
	cd libsass && make
	cd jsmin && make

golib:
	go get github.com/BurntSushi/toml
	go get bitbucket.org/kardianos/osext

clean:
	cd libsass && make clean
	cd jsmin && make clean

