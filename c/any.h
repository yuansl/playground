#ifndef ANY_H_X
#define ANY_H_X

#include <string.h>

struct type {
	const char *name;
	size_t size;
	size_t align;
	enum {
		CHAR,
		SHORT,
		INT,
		LONG,
		LLONG,
		UCHAR,
		USHORT,
		UINT,
		ULONG,
		ULLONG,
		FLOAT,
		DOUBLE,
		CHAR_PTR,
		UCHAR_PTR,
	} kind;
};

struct any {
	struct type type;
	char value[];
};

typedef struct any *any_t;

#define KIND_OF(x) \
	_Generic(x,                         \
		int: INT,                   \
		long: LONG,                 \
		unsigned: UINT,             \
		double: DOUBLE,             \
		char *: CHAR_PTR,           \
		const char *: CHAR_PTR,     \
		unsigned char *: UCHAR_PTR, \
		const unsigned char *: UCHAR_PTR)

extern const char *kind_name[];

#define SIZEOF(x)                           \
	({                                  \
		typeof(x) _tmp = (x);       \
		size_t _size;               \
		_size = _Generic(_tmp,				\
                        char * : strlen((const char *)&_tmp),	\
                        default : sizeof(_tmp)); \
		_size;                      \
	})

#define ANY(x)                                                                 \
	({                                                                     \
		size_t size = SIZEOF(x);                                       \
		any_t val = malloc(sizeof(*val) + (size + 7) / 8 * 8);         \
		val->type = (struct type){                                     \
			.name = kind_name[KIND_OF(x)],                         \
			.size = size,                                          \
			.align = alignof(x),                                   \
			.kind = KIND_OF(x),                                    \
		};                                                             \
		typeof(x) _tmp = x;                                            \
		char *_a = _Generic(x, char * : _tmp, default : (char*)&_tmp); \
		memcpy(val->value, _a, size);                                  \
		val;                                                           \
	})

void inspect_any(any_t val);

#endif
