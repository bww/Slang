
all:
	cd dep && make
	cd src && make

clean:
	cd src && make clean

cleanall: clean
	cd dep && make clean

