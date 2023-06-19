#include <fcntl.h>
#include <unistd.h>

#include "util.h"

int main(int argc, char *argv[])
{
	struct person *p;
	int fd;

	fd = open(argv[1], O_RDONLY);
	if (fd < 0)
		fatal("open error: %m\n");

	p = mmap(NULL, MMAP_SIZE_MAX, PROT_READ, MAP_SHARED, fd, 0);
	if (mmap_failed(p))
		fatal("mmap error: %m\n");

	close(fd);

	do {
		printf("person age = %d, gender = %d\n", p->age, p->gender);
		sleep(1);
	} while (p->age > 0 && p->gender > 0);

	return 0;
}
