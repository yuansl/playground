#ifndef UTIL_H_SOME
#define UTIL_H_SOME

#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

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

#define WITH_LOCKED(mutex, BODY)             \
	{                                    \
		pthread_mutex_lock(mutex);   \
		do {                         \
			BODY;                \
		} while (0);                 \
		pthread_mutex_unlock(mutex); \
	}

#define ARRAY_SIZE(a) (sizeof((a)) / sizeof((a)[0]))

static inline void init_rand(void)
{
	struct timespec time;
	clock_gettime(CLOCK_REALTIME, &time);

	srand(time.tv_nsec + time.tv_sec * 1e9);
}

#endif
