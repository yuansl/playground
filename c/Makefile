SHELL=/bin/bash
CC = gcc-13 -std=gnu2x
CFLAGS = -g -O2 -Wall -Wextra -fno-strict-aliasing

OBJS = main.o
OBJS += slice.o
OBJS += stringbuffer.o
OBJS += any.o

some: $(OBJS)
	$(CC) $(CFLAGS) -o $@ $(OBJS)
clean:
	$(RM) some $(OBJS)
