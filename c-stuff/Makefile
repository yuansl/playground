CC = gcc
CFLAGS =-std=gnu11 -Wall -Wextra -pthread -g -pipe
LDFLAGS = -pthread 

all: a.out

OBJS = main.o
OBJS += string_reader.o
OBJS += util.o
OBJS += printf.o

a.out: $(OBJS)
	$(CC) -o a.out $(LDFLAGS) $(OBJS)

clean:
	@rm -f *.o
