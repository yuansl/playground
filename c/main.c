#include <inttypes.h>
#include <sys/cdefs.h>
#define _GNU_SOURCE
<<<<<<< Updated upstream
#include <math.h>
=======
#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <assert.h>
>>>>>>> Stashed changes

#include "util.h"
#include "slice.h"
#include "stringbuffer.h"
#include "any.h"

#define SLICE_SIZE_MAX 128

<<<<<<< Updated upstream
#define ARRAY_SIZE(a) (sizeof((a)) / sizeof((a[0])))

[[maybe_unused]] static void test_stringbuffer(void)
{
	stringbuffer_t *array = create_buffer(SLICE_SIZE_MAX);
=======
_Noreturn static inline void _fatal(const char *fmt, ...)
{
	va_list ap;

	va_start(ap);
	vfprintf(stderr, fmt, ap);
	va_end(ap);

	exit(1);
}

#define fatal(...)                                                            \
	do {                                                                  \
		fprintf(stderr, "%s:%s:%d fatal error: ", __FILE__, __func__, \
			__LINE__);                                            \
		_fatal(__VA_ARGS__);                                          \
	} while (false)

typedef unsigned char byte;

typedef struct slice {
	size_t cap;
	size_t size;
	byte data[];
} slice_t;

static slice_t *slice_create(size_t cap)
{
	slice_t *buf = malloc(sizeof(struct slice) + cap);
	assert(buf != NULL);
	buf->cap = cap;
	buf->size = 0;
	return buf;
}

static void slice_destroy(slice_t *slice)
{
	free(slice);
}

static inline size_t slice_available(struct slice *slice)
{
	return slice->cap - slice->size;
}

int slice_append(struct slice *slice, const byte *bytes)
{
	size_t len = strlen((const char *)bytes);

	if (len > slice_available(slice)) {
		return ENOSPACE;
	}
	memcpy(slice->data, bytes, len);
	slice->size += len;

	return 0;
}

byte *slice_bytes(struct slice *slice, int at, size_t nbytes)
{
	static byte buf[BUFSIZ];

	if (slice->size == 0 || at < 0 || at >= (int)slice->size) {
		return NULL;
	}
	if (slice->size > sizeof(buf)) {
		fatal("BUG: sizeof(buf) too small, maybe you should grow it");
	}
	memcpy(buf, slice->data + at, nbytes);

	return buf;
}

typedef struct {
	slice_t *buf;		 /* buffer */
	const char *description; /* description */
	size_t w_off;		 /* write offset */
	size_t r_off;		 /* read offset */
} stringbuffer_t;

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

enum { STATE = 1L << 32 };

int main(void)
{
	stringbuffer_t *array = create_buffer(ARRAY_SIZE);
>>>>>>> Stashed changes
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

<<<<<<< Updated upstream
struct iterator {
	void *begin, *end;
	void *pos;
};
=======
	constexpr int SIZE = 100;

	struct {
		/* empty */
	} empty_structs[SIZE];
>>>>>>> Stashed changes

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
