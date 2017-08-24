all:
	cd tcpdump && make
	cd protocol && make

clean:
	@rm -f tcpdump/dumper protocol/protocol tcpdump/*.tcpdump tcpdump/*.txt
