#include <stddef.h>
#include <stdio.h>
#include <string.h>
#include <stdint.h>

int main(int argc, char *argv[])
{
	int a[16] = {0};
	int *p0;
	int *p1;
	ptrdiff_t diff;

	p0 = &a[2];
	p1 = (void *)((uintptr_t)a + (2 * sizeof(int)));
	diff = (uintptr_t)&a[2] - (uintptr_t)&a[1];
	printf("diff = %ld, p0=%p, a+2=%p, p0==p1:%s\n", diff, p0, p1,p0==p1?"true":"false");
	
	return 0;
}
