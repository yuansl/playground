#define _GNU_SOURCE
#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

#define ARRAY_SIZE 128

enum {
	ENOSPACE = -2, /* there is no space in buffer */
	EUNAVAILABLE   /* something went wrong, maybe a bug */
};

__attribute__((noreturn)) static inline void __fatal(const char *fmt, ...)
{
	va_list ap;

	vfprintf(stderr, fmt, ap);
	va_end(ap);

	exit(1);
}

#define fatal(...)                                                            \
	do {                                                                  \
		fprintf(stderr, "%s:%s:%d fatal error: ", __FILE__, __func__, \
			__LINE__);                                            \
		__fatal(__VA_ARGS__);                                         \
	} while (false)

typedef unsigned char byte;

typedef struct slice {
	size_t cap;
	size_t size;
	byte data[];
} slice_t;

slice_t *slice_create(size_t cap)
{
	slice_t *buf = malloc(sizeof(struct slice) + cap);
	buf->cap = cap;
	buf->size = 0;
	return buf;
}

void slice_destroy(slice_t *slice)
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

int main()
{
	stringbuffer_t *array = create_buffer(ARRAY_SIZE);
	const char *greet = "你好"; /* ,world */
	struct {
		short value;
	} x = { .value = 0xfefa };
	const char *p = (const char *)&x;

	for (size_t i = 0; i < sizeof(x); i++) {
		uint8_t x = p[i];
		printf("p[%zd]= %s\n", i, x < 0 ? "true" : "false");
	}

	int8_t y = -128;
	int8_t y0 = 0b11111010;

	printf("y = %#hhb, y0=%d\n", y, y0);

	if (buffer_append(array, (const byte *)greet) < 0) {
		fatal("BUG: buffer_append(%s):", greet);
	}

	printf("msg='%s'\n", buffer_bytes(array));

	if (buffer_append(array, (const byte *)"this is another message") < 0) {
		fatal("buffer_append error: out of memory");
	}

	printf("after append new message, now msg='%s'\n", buffer_bytes(array));

	buffer_destroy(array);

	struct {
		/* empty */
	} empty_structs[100];

	printf("sizeof(empty_structs)=%zd\n", sizeof empty_structs);

	return 0;
}
