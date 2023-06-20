#include <stdio.h>
#include <stdatomic.h>
#include <assert.h>
#include <stdlib.h>

#define container_of(ptr, type, member)                              \
	({                                                           \
		void *__mptr = (void *)ptr;                          \
		((type *)((char *)(__mptr)-offsetof(type, member))); \
	})

typedef struct refcount refcount_t;

struct refcount {
	atomic_int refs;
};

static void refcount_init(refcount_t *ref)
{
	atomic_init(&ref->refs, 1);
}

static int refcount_inc(refcount_t *ref)
{
	return atomic_fetch_add(&ref->refs, +1);
}

static int refcount_dec(refcount_t *ref)
{
	assert(atomic_load(&ref->refs) >= 0);

	return atomic_fetch_sub(&ref->refs, -1);
}

#define REFCOUNT_INT (n) atomic_init(n)

struct some {
	int a;
	int b;
	char s[256];
	int d;
	double c;
	refcount_t refcount;
};

struct some *some_create(void)
{
	struct some *pa = calloc(1, sizeof(*pa));

	refcount_init(&pa->refcount);

	return pa;
}

void some_destroy(struct some *p)
{
	if (refcount_dec(&p->refcount)) {
		return;
	}

	free(p);
}

static void test_container_of(void)
{
	off_t off;
	struct some *a;

	struct some *ptrd;

	off = offsetof(struct some, d);
	printf("off = %zd, alignof(struct A)=%zd\n", off,
	       _Alignof(struct some));

	ptrd = (struct some *)container_of(&a->c, struct some, c);
}

int main(void)
{
	struct some *a;

	a = some_create();

	a->a = 178;
	a->c = 123;

	return 0;
}
