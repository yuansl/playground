#include <stdio.h>
#include <stdint.h>

union x {
	int i;
	double d;
};

int foo(union x x)
{
	x.d = 0.31;
	return x.i;
}

uint32_t swap(uint32_t arg)
{
	uint16_t *sp = (uint16_t *)&arg;
	uint16_t hi, lo;
	uint16_t *p;

	hi = sp[0];
	lo = sp[1];

	sp[1] = hi;
	sp[0] = lo;

	p = (uint16_t *)&arg;
	*p = 3;

	return arg;
}

uint16_t foo2(uint64_t *x)
{
	uint16_t y = *(uint16_t *)x;
	return y;
}

uint16_t accumulate(uint32_t x)
{
	uint16_t sum = 0;

	for (int i = 0; i < 20; i++) {
		sum += (uint16_t)x;
	}
	return sum;
}

int a = 5;
int *b = &a;

int foo3(double *b)
{
	a = 12;
	*b = 0.314;

	return a;
}

int main(int argc, char *argv[])
{
	uint32_t val;
	union x x = {0};
	uint64_t y;

	val = swap((int32_t)0x01020304);
	printf("x = %#x\n", val);

	printf("x = %d\n", foo(x));

	printf("accumulate = %u\n", accumulate((uint32_t)0x01020304));

	y = 0x0102030405060708;
	foo2(&y);

	printf("a = %d\n", foo3((double *)&a));
	
	return 0;
}
