SHELL=/bin/bash
CC=gcc
CPPFLAGS=
CFLAGS=-O2 -g -pipe
LIBS=$(shell pkg-config --libs wayland-client)

all: a.out

a.out: main.o
	$(CC) -o $@ $< $(LIBS)

clean:
	@$(RM) main.o
