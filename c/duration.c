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

time_t duration_seconds(duration_t d)
{
	return d / SECOND;
}

int duration_minutes(duration_t d)
{
	return d / MINUTE;
}

double duration_hours(duration_t d)
{
	return (double)d / HOUR;
}

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
	duration_t duration = 365 * 24 * HOUR;
	bool gt = duration > 360 * 24 * HOUR;

	(void)gt;

	printf("seconds=%lds, minutes=%dm, hour=%.2fh\n",
	       duration_seconds(duration), duration_minutes(duration),
	       duration_hours(duration));

	return 0;
}
