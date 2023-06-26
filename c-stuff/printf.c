#include <stdarg.h>
#include <stddef.h>
#include <stdio.h>
#include <string.h>
#include <stdint.h>

#define swap(a, b)                   \
	({                           \
		typeof(a) __tmp = a; \
		(a) = (b);           \
		(b) = __tmp;         \
	})

static void reverse_string(char s[])
{
	for (int i = 0, j = strlen(s) - 1; i < j; i++, j--) {
		swap(s[i], s[j]);
	}
}

static int least_bit_pattern_of(long value, char buf[], size_t size)
{
	char *p = buf;

	if (value <= 0) {
		*p++ = '0';
	} else {
		while (value > 0) {
			if ((size_t)(p - buf) >= size) {
				return -1;
			}
			*p++ = value % 10 + '0';
			value /= 10;
		}
	}
	*p = '\0';

	return 0;
}

static int strconv(int value, char buf[], size_t size)
{
	least_bit_pattern_of(value, buf, size);

	reverse_string(buf);

	return 0;
}

static int strconvf(double value, char buf[], size_t size)
{
	size_t n;

	if (size == 0) {
		return -1;
	}

	least_bit_pattern_of((long)value, buf, size);
	reverse_string(buf);

	n = strlen(buf);
	buf[n] = '.';
	n++;
	value -= (long)value;

	least_bit_pattern_of((long)(value * 1000), buf + n, size - n);
	reverse_string(buf + n);

	return 0;
}

#define MAX_BITS 20

int Printf(const char *format, ...)
{
	int n;
	char buf[MAX_BITS];
	va_list ap;

	va_start(ap, format);
	n = 0;
	while (*format) {
		if (*format != '%') {
			putchar(*format);
		} else {
			format++;
			switch (*format) {
			case '%':
				putchar('%');
				break;
			case 'd': {
				int v = va_arg(ap, int);
				strconv(v, buf, sizeof(buf));
				fputs(buf, stdout);
				break;
			}
			case 'f': {
				double f = va_arg(ap, double);
				strconvf(f, buf, sizeof(buf));
				fputs(buf, stdout);
				break;
			}
			case 'c': {
				char c = va_arg(ap, int);
				putchar(c);
				break;
			}
			case 's': {
				char *s = va_arg(ap, char *);
				fputs(s, stdout);
				break;
			}
			default:
				fputs("Unknown type: ", stdout);
				putchar(*format);
				return -1;
			}
		}

		format++;
	}
	va_end(ap);

	return n;
}
