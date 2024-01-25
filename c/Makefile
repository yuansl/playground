SHELL=/bin/bash
CC = gcc-13 -std=gnu2x
CFLAGS = -g -O2 -Wall -Wextra -fno-strict-aliasing -pipe

OBJS = main.o
OBJS += slice.o
OBJS += stringbuffer.o
OBJS += any.o
OBJS += pthreads.o

a.out: $(OBJS)
	$(CC) $(CFLAGS) -o $@ $(OBJS)
clean:
	$(RM) a.out $(OBJS)
