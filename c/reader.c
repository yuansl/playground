#include <sys/stat.h>
#include <stdio.h>

#include "util.h"
#include "reader.h"

ssize_t consume_reader(Reader r, char *buf, size_t size)
{
	ssize_t nbytes;

	if ((nbytes = r->ops.read(r, buf, size)) < 0) {
		return -1;
	}
	return nbytes;
}
