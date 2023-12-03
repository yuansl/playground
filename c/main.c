#define _GNU_SOURCE
#include <stdio.h>
#include <stdarg.h>

void foo(...)
{
	va_list ap;

	va_start(ap);
	do {
		char *x = va_arg(ap, char *);
		if (!x) {
			break;
		}
		printf("x = %s\n", x);
	} while (false);
	va_end(ap);
}

typedef unsigned char byte;

int main(int argc, char *argv[])
{
	const byte *msg = u8 "hello";
	int x = 0b1001;

	const byte *c = u8 "some";
	printf("msg='%s', c = '%s'\n", msg, c);

	foo("some", "thing", "went", "wrong", nullptr);
	return 0;
}
