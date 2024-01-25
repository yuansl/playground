#include <stdio.h>
#include "util.h"

#define NR_CASES 2

struct iter {
	int i;
	int size;
	void *slice;
};

static int next(struct iter *iter)
{
	return ((int *)iter->slice)[iter->i++];
}

static void print_iter(struct iter *it)
{
	for (int i = 0; i < it->size; i++) {
		int tmp = next(it);
		printf("a[%d]=%d\n", i, tmp);
	}
}

int main(void)
{
	int a[] = { 1, 2, 3, 4 };
	int b[] = { 6, 7, 8 };
	struct {
		int *array;
		int size;
	} cases[NR_CASES] = {
		{ .array = a, .size = ARRAY_SIZE(a) },
		{ .array = b, .size = ARRAY_SIZE(b) },
	};

	for (int i = 0; i < NR_CASES; i++) {
		struct iter it = {
			.i = 0,
			.size = cases[i].size,
			.slice = cases[i].array,
		};
		print_iter(&it);
	}
	return 0;
}
