#ifndef READER_H
#define READER_H

#include <features.h>
#include <stddef.h>

__BEGIN_DECLS

#define container_of(ptr, type, member)                     \
	({                                                  \
		void *_mptr = (void *)ptr;                  \
		((type *)(_mptr - offsetof(type, member))); \
	})

struct reader;
struct reader_operations {
	int (*read)(struct reader *, void *buf, size_t size);
	int (*write)(struct reader *, const char *buf, size_t size);
};

struct reader {
	void *data;
	const struct reader_operations *r_ops;
};

typedef struct reader *Reader;

Reader new_string_reader(const char *buf);

__END_DECLS

#endif
