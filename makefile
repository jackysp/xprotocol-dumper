all:
	cd tcpdump && make
	cd protocol && make

clean:
	@rm -rf *.tcpdump *.txt
