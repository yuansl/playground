#include <time.h>
#include <stdio.h>

typedef enum : time_t {
	NANOSECOND = 1,
	MICROSECOND = NANOSECOND * 1000,
	MILLSECOND = MICROSECOND * 1000,
	SECOND = MILLSECOND * 1000,
	MINUTE = 60 * SECOND,
	HOUR = 60 * MINUTE,
} duration_t;

void pretty_print_duration(duration_t duration)
{
	duration_t seconds = duration / (time_t)1e9;
	duration_t nanoseconds = duration % (time_t)1e9;

	if (seconds > 0) {
		duration_t minutes = seconds / MINUTE;
		seconds %= MINUTE;
		printf("minutes=%ld\n", minutes);
	}

	printf("seconds=%ld, nanoseconds=%ld\n", seconds, nanoseconds);
}

int main(void)
{
	duration_t duration = 30 * MINUTE;

	pretty_print_duration(duration);

	return 0;
}
