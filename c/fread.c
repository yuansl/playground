#include <stdio.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdint.h>

#include "util.h"

#ifndef NAME_MAX
#define NAME_MAX 255
#endif

enum gender : uint8_t { MALE, FEMALE };

struct person {
	char name[NAME_MAX];
	int age;
	enum gender gender;
};

void bluk_write(const char filename[static 1])
{
	FILE *fp;

	fp = fopen(filename, "a+");
	if (!fp)
		fatal("open: %m");

	fwrite(&(struct person){ .name = "liming", .age = 38, .gender = MALE },
	       1, sizeof(struct person), fp);

	fwrite(&(struct person){ .name = "lixue", .age = 39, .gender = FEMALE },
	       1, sizeof(struct person), fp);

	fwrite(&(struct person){ .name = "lixuechun",
				 .age = 31,
				 .gender = FEMALE },
	       1, sizeof(struct person), fp);
}

#define WITH_OPEN_WRITE(filename, f, BODY)          \
	{                                           \
		FILE *__fp = fopen(filename, "w+"); \
		if (!__fp)                          \
			fatal("fopen: %m");         \
		typeof(__fp) f = __fp;              \
		do {                                \
			BODY;                       \
		} while (0);                        \
		fclose(__fp);                       \
	}

#define WITH_OPEN_READ(filename, f, BODY)          \
	{                                          \
		FILE *__fp = fopen(filename, "r"); \
		if (!__fp)                         \
			fatal("fopen: %m");        \
		typeof(__fp) f = __fp;             \
		do {                               \
			BODY;                      \
		} while (0);                       \
		fclose(__fp);                      \
	}

void bulk_read(const char filename[static 1])
{
	WITH_OPEN_READ(filename, f, {
		do {
			struct person x = {};

			fread(&x, 1, sizeof(x), f);

			printf("name:%-20s, age:%d, gender:%s\n", x.name, x.age,
			       x.gender ? "MALE" : "FEMALE");
		} while (!feof(f));
	});
}

int main(void)
{
	bulk_read("/tmp/some");
	return 0;
}
