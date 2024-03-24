CC=gcc -std=gnu2x
CFLAGS=-O2 -Wall -Wextra -O2 -pipe -fPIC
# $(shell sudo chown root raise-emacsd && sudo chmod +s raise-emacsd)

all: raise-emacsd

raise-emacsd: main.o
	$(CC) $(CFLAGS) -o $@ $? $(LDFLAGS)

install: raise-emacsd
	$(shell sudo chown root raise-emacsd && sudo chmod +s raise-emacsd && sudo mv raise-emacsd /usr/local/bin)

clean:
	$(RM) *.o *.a *.so
