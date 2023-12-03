#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <errno.h>
#include <stdlib.h>

int main(int argc, char *argv[])
{
	int a[(1L<<31) - 1];
	int *b;
	unsigned int *c;

	b = malloc(sizeof(*b));
	c = malloc(sizeof(*c));

	printf("ptrdiff_max=%ld, size_max=%lu, b=%p, c=%p\n", PTRDIFF_MAX, SIZE_MAX, b, c);
	free(b);
	free(c);
	return 0;
}
