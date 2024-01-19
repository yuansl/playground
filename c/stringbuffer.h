#ifndef STRINGBUFFER_H_
#define STRINGBUFFER_H_

#include "slice.h"

typedef struct {
	slice_t *buf; /* buffer */
	const char *description; /* description */
	size_t w_off; /* write offset */
	size_t r_off; /* read offset */
} stringbuffer_t;

stringbuffer_t *create_buffer(size_t cap);
void buffer_destroy(stringbuffer_t *array);
size_t buffer_available(stringbuffer_t *array);
int buffer_append(stringbuffer_t *array, const byte *msg);
int buffer_read(stringbuffer_t *array, byte buf[], size_t size);
byte *buffer_bytes(stringbuffer_t *array);

#endif
