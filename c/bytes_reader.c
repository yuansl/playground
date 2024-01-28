#include <string.h>
#include <stdio.h>

#include "reader.h"

struct bytes_reader {
	char buf[BUFSIZ];
};

static ssize_t read_from_string(struct reader *ctx, void *buf, size_t size)
{
	struct bytes_reader *r = ctx->data;
	size_t nbytes	       = strlen(r->buf);

	if (nbytes <= 0 || size <= 0) {
		return 0;
	}

	if (nbytes > size)
		nbytes = size;

	strncpy(buf, r->buf, size);

	return nbytes;
}

void test_consume_reader(void)
{
	struct bytes_reader reader = {
		.buf = "hello, this message comes from bytes_reader"
	};
	char buf[BUFSIZ];
	ssize_t nbytes;

	nbytes = consume_reader(
		&(struct reader){
			.data = &reader,
			.ops  = { .read = read_from_string },
		},
		buf, sizeof(buf));
	if (nbytes < 0)
		fatal("consume_reader(bytes_reader): %m");

	printf("Read %zd bytes from bytes_reader: '%s'\n", nbytes, buf);
}
