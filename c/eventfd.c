#include <sys/eventfd.h>
#include <unistd.h>
#include <err.h>
#include <inttypes.h>
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char *argv[])
{
	int efd;
	eventfd_t u;
	ssize_t s;
	pid_t pid;

	if (argc < 2) {
		fprintf(stderr, "Usage: %s <num>...\n", argv[0]);
		exit(EXIT_FAILURE);
	}
	efd = eventfd(0, 0);
	if (efd == -1)
		err(EXIT_FAILURE, "eventfd");

	pid = fork();
	if (pid < 0)
		err(EXIT_FAILURE, "fork");
	if (pid == 0) {
		for (int i = 1; i < argc; i++) {
			printf("Child %d writing %s to efd\n", getpid(),
			       argv[i]);

			/* strtoull() allows various bases */
			u = strtoull(argv[i], NULL, 0);

			s = eventfd_write(efd, u);
			if (s != sizeof(eventfd_t))
				err(EXIT_FAILURE, "write");
		}
		printf("Child %d completed write loop\n", getpid());

		exit(EXIT_SUCCESS);
	}
	sleep(2);

	printf("Parent %d about to read\n", getpid());

	u = 0;
	s = eventfd_read(efd, &u);
	if (s != sizeof(eventfd_t))
		err(EXIT_FAILURE, "read");

	printf("Parent %d read %" PRIu64 " (%#" PRIx64 ") from efd\n", getpid(),
	       u, u);

	return 0;
}
