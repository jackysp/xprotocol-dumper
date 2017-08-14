dumper: dumper.c tcpip.h
	gcc -O2 -Wall $^ -o $@ -lpcap
	chmod 755 dumper

clean:
	@rm -rf *.o dumper
