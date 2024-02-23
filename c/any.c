#include <stdio.h>
#include <string.h>

#include "util.h"
#include "any.h"

const char *kind_name[] = {
	[INT]	   = "int",
	[LONG]	   = "long",
	[DOUBLE]   = "double",
	[CHAR_PTR] = "char *",
};

char *stringify(any_t val)
{
	static char buf[1024];
	switch (val->type.kind) {
	case INT:
	case UINT: {
		int *value = (int *)val->value;
		snprintf(buf, sizeof(buf), "%d", (int)*value);
		break;
	}
	case DOUBLE: {
		double *value = (double *)val->value;
		snprintf(buf, sizeof(buf), "%.4f", (double)*value);
		break;
	}
	case LONG:
	case ULONG: {
		long *value = (long *)val->value;
		snprintf(buf, sizeof(buf), "%lu", (unsigned long)*value);
		break;
	}
	case CHAR_PTR:
	case UCHAR_PTR: {
		byte *value = val->value;
		strncpy(buf, (const char *)value, sizeof(buf) - 1);

		break;
	}

	default:
		panic();
	}
	return buf;
}

void inspect_any(any_t val)
{
	printf("type: %s, size: %zd, align: %zd", val->type.name,
	       val->type.size, val->type.align);

	switch (val->type.kind) {
	case INT:
	case UINT: {
		int *value = (int *)val->value;

		printf(", value = %d\n", *value);
		break;
	}
	case DOUBLE: {
		double *value = (double *)val->value;

		printf(", value = %.2f\n", *value);
		break;
	}
	case LONG:
	case ULONG: {
		long *value = (long *)val->value;

		printf(", value = %ld\n", *value);
		break;
	}
	case CHAR_PTR:
	case UCHAR_PTR: {
		const char *value = (const char *)val->value;

		printf(", value = %s\n", value);
		break;
	}

	default:
		panic();
	}
}
