CC=gcc-13 -std=gnu2x
CFLAGS = -O2 -Werror -Wall -Wextra
CPPFLAGS = $(shell pkg-config -cflags dbus-1)
LDFLAGS = -s -pie
LIBS = $(shell pkg-config --libs dbus-1) $(shell pkg-config --libs jansson)
EXECUTABLE = raise-emacs

all: $(EXECUTABLE)

$(EXECUTABLE): main.o
	$(CC) -o $@ $< $(LDFLAGS) $(LIBS)

clean:
	$(RM) *.o $(EXECUTABLE)
