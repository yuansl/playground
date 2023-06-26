#include <sys/time.h>
#include <stddef.h>

void gettime(void)
{
	struct timeval tv;

	gettimeofday(&tv, NULL);
}
