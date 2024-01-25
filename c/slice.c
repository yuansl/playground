#include <stdlib.h>
#include <string.h>

#include "util.h"
#include "slice.h"

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

size_t slice_available(slice_t *slice)
{
	return slice->cap - slice->size;
}

int slice_append(slice_t *slice, const byte *bytes)
{
	size_t len = strlen((const char *)bytes);

	if (len > slice_available(slice)) {
		return ENOSPACE;
	}

	memcpy(slice->data, bytes, len);
	slice->size += len;

	return 0;
}

#define BUFSIZE 4096

byte *slice_bytes(slice_t *slice, int at, size_t nbytes)
{
	static byte buf[BUFSIZE];

	if (slice->size == 0 || at < 0 || at >= (int)slice->size) {
		return NULL;
	}
	if (slice->size > sizeof(buf)) {
		fatal("BUG: sizeof(buf) too small, maybe you should grow it");
	}

	memcpy(buf, slice->data + at, nbytes);

	return buf;
}
