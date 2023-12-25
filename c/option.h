#ifndef _OPTION
#define _OPTION

typedef struct option *option_t;

struct option {
	void (*apply)(option_t, void *options);
	void *data;
};

void option_destroy(option_t opt);

#endif
