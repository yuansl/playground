#include <stdio.h>
#include <string.h>

void fatal(const char *format, ...);

struct reader_context {
	void *data;
};

typedef struct reader {
	struct reader_context *ctx;
	ssize_t (*Read)(struct reader_context *ctx, void *buf, size_t size);
} Reader;

void consume(Reader *r)
{
	char buf[BUFSIZ];
	ssize_t nbytes;

	if ((nbytes = r->Read(r->ctx, buf, sizeof(buf))) < 0) {
		fatal("reader.read error");
		return;
	}

	buf[nbytes] = '\0';

	printf("Read %zd bytes: %s\n", nbytes, buf);
}

struct bytes_reader {
	char buf[BUFSIZ];
};

ssize_t __attribute__((visibility("hidden")))
read_from_string(struct reader_context *ctx, void *buf, size_t size)
{
	struct bytes_reader *r = ctx->data;
	size_t nbytes = strlen(r->buf);

	if (nbytes <= 0 || size <= 0) {
		return 0;
	}

	strcpy(buf, r->buf);

	return nbytes;
}
