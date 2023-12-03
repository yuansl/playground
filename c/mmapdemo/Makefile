all: mmapr mmapw

mmapw: main.o
	$(CC) -o $@ $<

mmapr: mmapr.o
	$(CC) -o $@ $<

clean: *.o
	rm -f $< mmapr mmapw
