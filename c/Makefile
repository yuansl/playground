SHELL=/bin/bash
CC = gcc-13 -std=gnu2x
CFLAGS = -g -O2 -Wall -Wextra -fno-strict-aliasing
LDFLAGS = $(shell pkg-config --libs jansson)
OBJS = main.o
OBJS += slice.o
OBJS += stringbuffer.o
OBJS += any.o
OBJS += pthreads.o
OBJS += option.o

a.out: $(OBJS)
	$(CC) $(CFLAGS) -o $@ $(OBJS) $(LDFLAGS)
clean:
	$(RM) a.out $(OBJS)
