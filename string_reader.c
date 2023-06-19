#include <sys/types.h>

#include <string.h>
#include <stdlib.h>
#include <assert.h>
#include <stdio.h>

#include "reader.h"

struct string_reader {
	char *buf;
	size_t len, cap;
	size_t off;
	struct reader i_reader;
};

int read_from_string(struct reader *r, void *buf, size_t size)
{
	struct string_reader *sr;
	ssize_t remain;

	sr = container_of(r, struct string_reader, i_reader);

	remain = sr->len - sr->off;
	if (remain <= 0) {
		return 0;
	}
	if (size > remain) {
		size = remain;
	}
	memcpy(buf, sr->buf + sr->off, size);
	sr->off += size;

	return size;
}

static int write_to_string(struct reader *r, const char *buf, size_t size)
{
	struct string_reader *sr;
	size_t remain;

	sr = container_of(r, struct string_reader, i_reader);
	remain = sr->cap - sr->len;
	if (size > remain) {
		if (size > sr->cap * 2) {
			sr->cap = size;
		} else {
			sr->cap *= 2;
		}
		sr->cap += sr->len;
		sr->buf = realloc(sr->buf, sr->cap);
		if (!sr->buf) {
			return -1;
		}
	}
	memcpy(sr->buf + sr->len, buf, size);
	sr->len += size;

	return size;
}

static const struct reader_operations string_reader_ops = {
	.read = read_from_string,
	.write = write_to_string,
};

#define BUFSIZE_MIN 16

Reader new_string_reader(const char *buf)
{
	struct string_reader *r;
	size_t size;

	assert(buf != NULL);

	r = calloc(1, sizeof(*r));
	r->i_reader.r_ops = &string_reader_ops;

	size = strlen(buf);
	if (size > 0) {
		r->cap = size * 2;
	} else {
		r->cap = BUFSIZE_MIN;
	}
	r->buf = malloc(r->cap);
	memcpy(r->buf, buf, size);
	r->len = size;

	return &r->i_reader;
}

/*
 * .. c:function:: int read_message (struct reader *r, void *buf, size_t size)
 *
 *    read message from **reader**
 *
 * **Parameters**
 *
 * ``struct reader *r``
 *   reader interface
 *
 * ``void *buf``
 *   buffer to read contents from reader **r**
 *
 * ``size_t size``
 *   size of **buf**
 *
 * **Return**
 *
 * return size of bytes read from **r**
 */
static inline int read_message(struct reader *r, void *buf, size_t size)
{
	return r->r_ops->read(r, buf, size);
}

static inline int write_message(struct reader *r, const char *buf, size_t size)
{
	return r->r_ops->write(r, buf, size);
}

#define FOREACH_STRING(strs) \
	for (typeof(&strs[0]) ITER = &strs[0]; *ITER != NULL; ITER = ITER + 1)

void test_string_reader(void)
{
	Reader r;
	char buf[BUFSIZ];
	const char *message = "hello, world";
	int n;

	r = new_string_reader("");

	write_message(r, message, strlen(message));

	static const char *strs[] = {
		"what ever1", "what ever2", "what ever3", "what ever4",
		"what ever5", "what ever6", "what ever7", NULL,
	};

	FOREACH_STRING(strs)
	{
		write_message(r, *ITER, 10);
	}

	for (int i = 0; i < 10; i++) {
		n = read_message(r, buf, 10);
		buf[n] = '\0';

		printf("#%d: read_message: buf = `%s`\n", i, buf);
	}
}
