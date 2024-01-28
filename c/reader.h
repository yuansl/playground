#ifndef READER_H
#define READER_H

#include <stdlib.h>
#include <stddef.h>

#include "util.h"

struct reader;

struct reader_operations {
	ssize_t (*read)(struct reader *ctx, void *buf, size_t size);
};

struct reader {
	void *data;
	struct reader_operations ops;
};

typedef struct reader *Reader;

ssize_t consume_reader(Reader r, char *buf, size_t size);

Reader new_string_reader(const char *buf);

#endif
