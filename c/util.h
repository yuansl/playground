#ifndef UTIL_H_SOME
#define UTIL_H_SOME

#include <stdarg.h> /* for va_{start,end} */
#include <stdio.h>
#include <stdlib.h>
#include <time.h> /* for clock_gettime */

#define NORETURN __attribute__((noreturn))
#define __unused __attribute__((unused))

typedef unsigned char byte;

#define ARRAY_SIZE(a) (int)(sizeof((a)) / sizeof((a)[0]))

#ifndef max
#define max(a, b)                  \
	({                         \
		typeof(a) _a = a;  \
		typeof(b) _b = b;  \
		_a > _b ? _a : _b; \
	})
#endif

#ifndef min
#define min(a, b)                  \
	({                         \
		typeof(a) _a = a;  \
		typeof(b) _b = b;  \
		_a < _b ? _a : _b; \
	})
#endif

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

#define WITH_LOCKED(mutex, BODY)             \
	{                                    \
		pthread_mutex_lock(mutex);   \
		do {                         \
			BODY;                \
		} while (0);                 \
		pthread_mutex_unlock(mutex); \
	}

#define WITH_OPEN_AS(filename, fp, BODY)         \
	{                                        \
		FILE *fp = fopen(filename, "r"); \
		do {                             \
			BODY;                    \
		} while (0);                     \
		fclose(fp);                      \
	}

/* strequal return true if two strings s1 and s2 are equal */
#define strequal(s1, s2) (strcmp(s1, s2) == 0)

static inline void init_rand(void)
{
	struct timespec time;

	clock_gettime(CLOCK_REALTIME, &time);
	srand(time.tv_nsec + time.tv_sec * 1e9);
}

#endif
