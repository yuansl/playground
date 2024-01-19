#include <stdlib.h>
#include <string.h>

#include "stringbuffer.h"

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
