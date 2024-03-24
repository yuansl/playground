SHELL=/bin/bash
CC = gcc-13 -std=gnu2x
CFLAGS = -O2 -Wall -Wextra -Werror -fno-strict-aliasing -pthread
LDFLAGS = -pthread
LIBS = $(shell pkg-config --libs jansson)
OBJS = main.o
OBJS += slice.o
OBJS += stringbuffer.o
OBJS += any.o
OBJS += pthreads.o
OBJS += option.o

EXECUTABLE = a.out

$(EXECUTABLE): $(OBJS)
	$(CC) -o $@ $(OBJS) $(LDFLAGS) $(LIBS)
clean:
	$(RM) $(EXECUTABLE) $(OBJS)
