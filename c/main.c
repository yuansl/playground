#include <inttypes.h>
#include <sys/cdefs.h>
#define _GNU_SOURCE
#include <math.h>

#include "util.h"
#include "slice.h"
#include "stringbuffer.h"
#include "any.h"

#define SLICE_SIZE_MAX 128

#define ARRAY_SIZE(a) (sizeof((a)) / sizeof((a[0])))

[[maybe_unused]] static void test_stringbuffer(void)
{
	stringbuffer_t *array = create_buffer(SLICE_SIZE_MAX);
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
typedef const char *string;

struct iterator {
	void *begin, *end;
	void *pos;
};

#define ITERATOR_INITIALIZER(a)                                  \
	{                                                        \
		.begin = a, .end = (a + ARRAY_SIZE(a)), .pos = a \
	}

#define zeroval(x) \
	_Generic(x,		      \
           char *: "",		      \
	   string: "",		      \
	   int: (int)NAN,	      \
	   unsigned: (unsigned)NAN,   \
	   double: NAN)

#define next(iter, T)                               \
	({                                          \
		void *it = (iter)->pos;             \
		(iter)->pos = (T *)(iter)->pos + 1; \
		typeof(T) x;                        \
		x = it ? *(T *)it : zeroval(x);     \
	})

static __attribute_maybe_unused__ void test_any(void)
{
	any_t values[] = { ANY(3.18), ANY(18), ANY("hello, world") };

	for (int i = 0; i < (int)ARRAY_SIZE(values); i++) {
		inspect_any(values[i]);
	}
}

union a_union {
	int i;
	double d;
};

int f(void)
{
	union a_union t;
	t.d = 3.0;
	return t.i;
}

int f2(void)
{
	union a_union t;
	int *ip;
	t.d = 3.0;
	ip = &t.i;
	return *ip;
}

int a = 3;

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

#define SIZE 16

int main(void)
{
	float *y[SIZE];

	y[0] = malloc(SIZE * sizeof(*y[0]));
	y[0][0] = 3.14;
	y[0][15] = 5.28;
	float x[SIZE][SIZE] = { { [0] = 3.14, [15] = 5.28 } };
	(void)x;
	matrix_fun(SIZE, y);

	{
		int num = 10;
		// Function Call
		int result = foo(&num, (long *)&num);
		// Print result
		printf("result %d\n", result);
	}
	{
		int x = change_a((double *)&a, &a);
		printf("x = %d, a = %d\n", x, a);
	}
	{
		union Some {
			long error;
			double result;
		} u;
		u.result = 10;
		printf("\u4E2D\u56FD u.error=%ld\n", u.error);
	}
	{
		printf("sizeof(true)=%zd,sizeof(bool)=%zd,sizeof(false)=%zd\n",
		       sizeof(true), sizeof(bool), sizeof(false));
	}
	/*
	 * int a[] = { 3, 8, 8, 9, 34, 838 };
	 * string b[] = { "whatever", "your", "want" };
	 * struct {
	 * 	void *slice;
	 * 	void *end;
	 * 	int size;
	 * } cases[] = { { a, a + ARRAY_SIZE(a), ARRAY_SIZE(a) },
	 * 	      { b, b + ARRAY_SIZE(b), ARRAY_SIZE(b) } };
	 * for (int i = 0; i < (int)ARRAY_SIZE(cases); i++) {
	 * 	struct iterator iter = {
	 * 		.begin = cases[i].slice,
	 * 		.end = cases[i].slice + cases[i].size,
	 * 		.pos = cases[i].slice,
	 * 	};
	 *
	 * 	for (int j = 0; j < cases[i].size; j++) {
	 * 		string x = next(&iter, string);
	 * 		printf("a[%d] = %s\n", j, x);
	 * 	}
	 * }
	 */
	return 0;
}
