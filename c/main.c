#define _GNU_SOURCE
#include <math.h>
#include <stdio.h>

#include "util.h"
#include "slice.h"
#include "stringbuffer.h"
#include "any.h"

constexpr int STRING_BUFSIZE = 10;

typedef const char *string;

struct iterator {
	void *begin, *end;
	void *pos;
};

#define ITERATOR_INITIALIZER(a)                                  \
	{                                                        \
		.begin = a, .end = (a + ARRAY_SIZE(a)), .pos = a \
	}

#define zeroval(x)                       \
	_Generic(x,                      \
		char *: "",              \
		string: "",              \
		int: (int)NAN,           \
		unsigned: (unsigned)NAN, \
		double: NAN)

#define next(iter, T)                               \
	({                                          \
		void *it = (iter)->pos;             \
		(iter)->pos = (T *)(iter)->pos + 1; \
		typeof(T) x;                        \
		x = it ? *(T *)it : zeroval(x);     \
	})

void __attribute_maybe_unused__ test_any(void)
{
	any_t values[] = { ANY(3.18), ANY(18), ANY("hello, world") };

	for (int i = 0; i < (int)ARRAY_SIZE(values); i++) {
		inspect_any(values[i]);
	}
}

union oneof {
	int i;
	double d;
};

int f(void)
{
	union oneof t;
	t.d = 3.0;
	return t.i;
}

int f2(void)
{
	union oneof t;
	int *ip;
	t.d = 3.0;
	ip = &t.i;
	return *ip;
}

int a = 3;

/* type aliasing */
int change_a(double *p, int *p2)
{
	int *x = (int *)p;

	*x = 42;

	return *p2;
}

int foo(int *ptr1, long *ptr2)
{
	*ptr1 = 10;
	*ptr2 = 11;

	return *ptr1;
}

void matrix_fun(const int N, const float x[N][N])
{
	printf("x[0][0]=%f\n", x[0][15]);
}

void test_stringbuffer(void)
{
	stringbuffer_t *array = create_buffer(STRING_BUFSIZE);
	const char *greet = "你好"; /* ,world */

	if (buffer_append(array, (const byte *)greet) < 0) {
		fatal("BUG: buffer_append(%s):", greet);
	}

	printf("msg='%s'\n", buffer_bytes(array));

	if (buffer_append(array, (const byte *)"this is another message") < 0) {
		fatal("buffer_append error: out of memory");
	}

	printf("after append new message, now msg='%s'\n", buffer_bytes(array));

	buffer_destroy(array);
}

int main(void)
{
	return 0;
}
