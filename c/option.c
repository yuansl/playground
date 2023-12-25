#include <stdlib.h>
#include "option.h"

void option_destroy(option_t opt)
{
	if (opt->data)
		free(opt->data);
	free(opt);
}
