#ifndef UTIL_H
#define UTIL_H

#define ARRAY_SIZE(x) (sizeof(x) / sizeof(x[0]))

#define lambda(statements, ...)       \
	({                            \
		void __f(__VA_ARGS__) \
		{                     \
			statements;   \
		};                    \
		__f;                  \
	})

#define for_each(list, func)                                      \
	({                                                        \
		for (int i = 0; i < (int)ARRAY_SIZE(list); i++) { \
			func(list[i]);                            \
		}                                                 \
	})
#define ARRAY_SIZE(x) (sizeof(x) / sizeof(x[0]))

#define lambda(statements, ...)       \
	({                            \
		void __f(__VA_ARGS__) \
		{                     \
			statements;   \
		};                    \
		__f;                  \
	})

#define for_each(list, func)                                      \
	({                                                        \
		for (int i = 0; i < (int)ARRAY_SIZE(list); i++) { \
			func(list[i]);                            \
		}                                                 \
	})

void fatal(const char *format, ...);

int Printf(const char *format, ...);
#endif
