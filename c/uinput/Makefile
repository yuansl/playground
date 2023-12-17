CC = gcc-12 -std=gnu2x
CFLAGS ?= -O2 -Wall -pipe
EXECUTABLE ?= raise-emacsd
PREFIX ?= /usr/local
BINDIR ?=$(PREFIX)/bin
OBJS = main.o

all: $(EXECUTABLE)

$(EXECUTABLE): $(OBJS)
	$(CC) -o $@ $< $(LDFLAGS)

clean:
	$(RM) $(EXECUTABLE) *.o

install:
	@install $(EXECUTABLE) $(BINDIR) && sudo chown root $(BINDIR)/$(EXECUTABLE) && sudo chmod +s $(BINDIR)/$(EXECUTABLE)
