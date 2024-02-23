#ifndef ANY_H_X
#define ANY_H_X

#include <stddef.h>

struct type {
	const char *name;
	size_t size;
	size_t align;
	enum {
		CHAR = 0,
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
		POINTER,
		STRUCT,
		UNION,
	} kind;
};

typedef unsigned char byte;

struct any {
	struct type type;
	byte value[];
};

typedef struct any *any_t;

#define KIND_OF(x)                                \
	_Generic(x,                               \
		int: INT,                         \
		long: LONG,                       \
		unsigned: UINT,                   \
		double: DOUBLE,                   \
		char *: CHAR_PTR,                 \
		const char *: CHAR_PTR,           \
		unsigned char *: UCHAR_PTR,       \
		const unsigned char *: UCHAR_PTR, \
		default: STRUCT)

extern const char *kind_name[];

#define SIZEOF(x)                                                 \
	({                                                        \
		typeof(x) __alias = x;                            \
		_Generic(__alias,                                 \
			const char *: strlen(*(char **)&__alias), \
			char *: strlen(*(char **)&__alias),       \
			default: sizeof(__alias));                \
	})

#define ANY(x)                                                               \
	({                                                                   \
		typeof(x) _alias_x = x;                                      \
		const char *_p	   = _Generic(_alias_x,                      \
			    char *: _alias_x,                                \
			    const char *: _alias_x,                          \
			    default: (char *)&_alias_x);                     \
		size_t __size	   = SIZEOF(x);                              \
		any_t __val = malloc(sizeof(*__val) + (__size + 7) / 8 * 8); \
		__val->type = (struct type){                                 \
			.name  = kind_name[KIND_OF(_alias_x)],               \
			.size  = __size,                                     \
			.align = alignof(_alias_x),                          \
			.kind  = KIND_OF(_alias_x),                          \
		};                                                           \
		memcpy(__val->value, _p, __size);                            \
		__val;                                                       \
	})

void inspect_any(any_t val);

char *stringify(any_t val);

#endif
