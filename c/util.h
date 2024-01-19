#ifndef UTIL_H_SOME
#define UTIL_H_SOME

#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>

#define NORETURN __attribute__((noreturn))

static inline NORETURN void __fatal(const char *fmt, ...)
{
	va_list ap;

	va_start(ap);
	vfprintf(stderr, fmt, ap);
	va_end(ap);

	exit(EXIT_FAILURE);
}

#define fatal(...)                                                          \
	do {                                                                \
		fprintf(stderr, "%s:%d fatal error: ", __FILE__, __LINE__); \
		__fatal(__VA_ARGS__);                                       \
	} while (false)

#define panic()                                                               \
	({                                                                    \
		fprintf(stderr, "BUG: %s:%d should be unreachable", __FILE__, \
			__LINE__);                                            \
		abort();                                                      \
	})

#endif
