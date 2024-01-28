#include <string.h>
#include <stdio.h>

#include "reader.h"

struct file_reader {
	FILE *fp;
};

static ssize_t read_from_file(struct reader *ctx, void *buf, size_t size)
{
	struct file_reader *reader = ctx->data;
	size_t nbytes		   = 0;

	while (!feof(reader->fp) && nbytes < size) {
		char buff[10];
		size_t n;
		char *s = fgets(buff, sizeof(buff), reader->fp);
		if (!s)
			break;
		n = strlen(s);
		if (nbytes + n > size)
			break;

		memcpy(buf + nbytes, buff, n);
		nbytes += n;
	}

	return nbytes;
}

void test_consume_reader2(void)
{
	const char *filename = "/tmp/some.c";

	WITH_OPEN_AS(filename, fp, {
		char buf[BUFSIZ];
		ssize_t nbytes;

		nbytes = consume_reader(
			&(struct reader){
				.data = &(struct file_reader){ .fp = fp },
				.ops  = { .read = read_from_file },
			},
			buf, sizeof(buf));
		if (nbytes < 0)
			fatal("consume_reader: %m");
		buf[nbytes] = '\0';
		printf("read %zd bytes from file_reader: '%s'\n", nbytes, buf);
	});
}
