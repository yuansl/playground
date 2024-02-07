#ifndef SLICE_H_
#define SLICE_H_

#include "util.h"

enum {
	ENOSPACE = -2, /* there is no space in buffer */
	EUNAVAILABLE /* something went wrong, maybe a bug */
};

typedef struct slice {
	size_t cap;
	size_t size;
	byte data[];
} slice_t;

slice_t *slice_create(size_t cap);
void slice_destroy(slice_t *slice);
int slice_append(slice_t *slice, const byte *bytes);
byte *slice_bytes(slice_t *slice, int at, size_t nbytes);
size_t slice_available(slice_t *slice);

#endif /*  SLICE_H_ */
