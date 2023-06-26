#include <unistd.h>

#include <stdio.h>
#include <fcntl.h>
#include <string.h>

#include "util.h"

#include "reader.h"

struct person {
	const char *name;
	int age;
	const char *address;
};

static void read_from_file(const char *filename)
{
	char buf[BUFSIZ];
	struct person *person = (struct person *)buf;
	int fd;

	fd = open(filename, O_RDONLY);
	if (fd < 0) {
		fatal("open: %m");
	}

	if (read(fd, person, sizeof(buf)) > 0) {
		printf("name=%s,age=%d,addres=%s\n", person->name, person->age,
		       person->address);
	}
}

static void write_to_file(struct person *liuming, const char *filename)
{
	int fd;

	fd = open(filename, O_CREAT | O_WRONLY, 0664);
	if (fd < 0) {
		fatal("open: %m");
	}
	if (write(fd, liuming, sizeof *liuming) != sizeof(*liuming)) {
		fatal("write: %m");
	}
	close(fd);
}

void test_read_write_file(void)
{
	const char *filename = "/tmp/something";
	char name[] = "Whatever";
	struct person liuxiang = {
		.name = name,
		.age = 26,
		.address = "Shanghai",
	};

	for (size_t i = 0; i < strlen(name); i++) {
		printf("%hhd\n", name[i]);
	}

	write_to_file(&liuxiang, filename);
	read_from_file(filename);
}

#define lambdax(return_type, arguments, body) \
	({ return_type __lambda arguments body __lambda; })

#define $(type, ...)        \
	(struct type)       \
	{                   \
		__VA_ARGS__ \
	}

struct point {
	float x, y;
};

#define foreach(x, slice)                        \
	typeof(slice[0]) x;                      \
	for (typeof(&slice[0]) __it = &slice[0]; \
	     __it < slice + ARRAY_SIZE(slice) && (x = __it); __it++)

#ifdef __CONCAT
#undef __CONCAT
#define __CONCAT1(x, y) x##y
#define __CONCAT(x, y) __CONCAT1(x, y)
#endif

int main(void)
{
	printf("3##4 = %d\n", __CONCAT(3, 4));
	printf("3+4##5+6 = %d\n", __CONCAT(3 + 4, 5 + 6));
	return 0;
}
