#define _GNU_SOURCE
#include <inttypes.h>
#include <math.h>
#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <assert.h>

#include "util.h"
#include "slice.h"
#include "stringbuffer.h"
#include "any.h"

#define SLICE_SIZE_MAX 128

constexpr int STRING_BUFSIZE = 10;

#define ARRAY_SIZE(a) (sizeof((a)) / sizeof((a[0])))

[[noreturn]] static inline void _fatal(const char *fmt, ...)
{
	va_list ap;

	va_start(ap);
	vfprintf(stderr, fmt, ap);
	va_end(ap);

	exit(1);
}

#define MAYBE_UNUSED [[maybe_unused]]

typedef unsigned char byte;

stringbuffer_t *create_buffer(size_t cap)
{
	stringbuffer_t *array = calloc(1, sizeof(*array));

	assert(array != NULL);

	array->buf = slice_create(cap);

	return array;
}

size_t buffer_available(stringbuffer_t *array)
{
	return slice_available(array->buf);
}

int buffer_append(stringbuffer_t *array, const byte *msg)
{
	int result;

	do {
		if ((result = slice_append(array->buf, msg)) < 0) {
			if (result == ENOSPACE) {
				int newcap = array->buf->cap * 2;
				array->buf = realloc(array->buf, newcap);
				assert(array != NULL);
				assert(array->buf != NULL);
				array->buf->cap = newcap;
				continue;
			}
			return result;
		}
	} while (false);

	array->w_off += strlen((const char *)msg);

	return 0;
}

int buffer_read(stringbuffer_t *array, byte buf[], size_t size)
{
	int result;
	byte *tmp;

	result = buffer_available(array);

	if (result < 0) {
		return result;
	} else if ((int)size > result) {
		return EUNAVAILABLE;
	}
	tmp = slice_bytes(array->buf, array->r_off, size);
	if (tmp) {
		memcpy(buf, tmp, size);
		array->r_off += size;
	}

	return 0;
}

byte *buffer_bytes(stringbuffer_t *array)
{
	return slice_bytes(array->buf, array->r_off,
			   array->w_off - array->r_off);
}

void buffer_destroy(stringbuffer_t *array)
{
	slice_destroy(array->buf);
	free(array);
}

typedef const char *string;

struct iterator {
	void *begin, *end;
	void *pos;
};

struct {
	/* empty */
} empty_structs[STRING_BUFSIZE];

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

int main(void)
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
