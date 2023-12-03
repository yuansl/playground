#include <asm-generic/errno-base.h>
#define _GNU_SOURCE

#include <unistd.h>
#include <string.h>
#include <limits.h>
#include <fcntl.h>
#include <errno.h>
#include "util.h"

static void defer_delete_file(int status, void *filename)
{
	unlink((const char *)filename);
}

static bool has_set_file_seals(int fd)
{
	int seals;

	seals = fcntl(fd, F_GET_SEALS);
	if (seals > 0) {
		if ((seals & F_SEAL_SEAL) == F_SEAL_SEAL)
			puts("Prevent further seals from being set");
		else if ((seals & F_SEAL_SHRINK) == F_SEAL_SHRINK)
			puts("Prevent further seals from shrinking");
		else if ((seals & F_SEAL_GROW) == F_SEAL_GROW)
			puts("Prevent further seals from growing");
		else if ((seals & F_SEAL_WRITE) == F_SEAL_WRITE)
			puts("Prevent further seals from writing");
		else if ((seals & F_SEAL_FUTURE_WRITE) == F_SEAL_FUTURE_WRITE)
			puts("Prevent future writes while mapped");

		return true;
	}

	return false;
}

static void *mmap_on_file(int fd, off_t offset, size_t size)
{
	void *addr;

	addr = mmap(NULL, size, PROT_WRITE, MAP_SHARED, fd, offset);
	if (mmap_failed(addr))
		fatal("mmap failed: %m\n");

	return addr;
}

int main(int argc, char *argv[])
{
	int fd;
	int seals;
	struct person *addr;

	if (argc < 3)
		fatal("Usage: %s <filename> <message>\n", argv[0]);

	fd = open(argv[1], O_CREAT | O_RDWR, 0666);
	if (fd < 0)
		fatal("open error: %d %m\n", errno);
	fallocate(fd, FALLOC_FL_ZERO_RANGE, 0, MMAP_SIZE_MAX);

	if (has_set_file_seals(fd)) {
		return 1;
	}

	addr = mmap_on_file(fd, 0, MMAP_SIZE_MAX);

	do {
		scanf("%d %d", &addr->age, &addr->gender);
	} while (addr->age >= 0 && addr->gender > 0);

	close(fd);

	return 0;
}
