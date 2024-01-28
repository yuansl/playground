#include <json-c/json_object.h>
#include <json-c/json_types.h>
#include <linux/limits.h>
#define _GNU_SOURCE
#include <math.h>
#include <stdio.h>

#include <json-c/json.h>

#include "util.h"
#include "slice.h"
#include "stringbuffer.h"
#include "any.h"

constexpr int STRING_BUFSIZE = 10;

/* maybe unused */
#define __unused __attribute__((unused))

typedef const char *string;

struct iterator {
	void *begin, *end;
	void *pos;
};

#define ITERATOR_INITIALIZER(a)                                   \
	{                                                         \
		.begin = a, .end = (a + ARRAY_SIZE(a)), .pos = a, \
	}

#define zeroval(x)                       \
	_Generic(x,                      \
		char *: "",              \
		string: "",              \
		int: (int)NAN,           \
		unsigned: (unsigned)NAN, \
		double: NAN)

#define next(iter, T)                               \
	({                                          \
		void *it    = (iter)->pos;          \
		(iter)->pos = (T *)(iter)->pos + 1; \
		typeof(T) x;                        \
		x = it ? *(T *)it : zeroval(x);     \
	})

__unused void test_any(void)
{
	any_t values[] = { ANY(3.18), ANY(18) }; //, ANY("hello, world") };

	for (int i = 0; i < (int)ARRAY_SIZE(values); i++) {
		inspect_any(values[i]);
	}
}

/* type aliasing */
union oneof {
	int i;
	double d;
};

int f(void)
{
	union oneof t;
	t.d = 3.0;
	return t.i;
}

int f2(void)
{
	union oneof t;
	int *ip;
	t.d = 3.0;
	ip  = &t.i;
	return *ip;
}

int a = 3;

int change_a(double *p, int *p2)
{
	int *x = (int *)p;

	*x = 42;

	return *p2;
}

int foo(int *ptr1, long *ptr2)
{
	*ptr1 = 10;
	*ptr2 = 11;

	return *ptr1;
}

void matrix_fun(const int N, const float x[N][N])
{
	printf("x[0][0]=%f\n", x[0][15]);
}

void test_stringbuffer(void)
{
	stringbuffer_t *array = create_buffer(STRING_BUFSIZE);
	const char *greet     = "你好"; /* ,world */

	if (buffer_append(array, (const byte *)greet) < 0) {
		fatal("BUG: buffer_append(%s):", greet);
	}

	printf("msg='%s'\n", buffer_bytes(array));

	if (buffer_append(array, (const byte *)"this is another message") < 0) {
		fatal("buffer_append error: out of memory");
	}

	printf("after append new message, now msg='%s'\n", buffer_bytes(array));

	buffer_destroy(array);
}

struct person {
	char name[NAME_MAX];
	int age;
	char blog[NAME_MAX];
	char addr[NAME_MAX];
	enum {
		MALE,
		FEMALE
	} gender;
};

static void person_pretty_print(struct person *p)
{
	printf("name: %s, age: %d, blog: '%s', addr: '%s'\n", p->name, p->age,
	       p->blog, p->addr);
}

void do_test_json_c(const char json[static 1], struct person *per)
{
	json_object *token = json_tokener_parse(json);

	json_object_object_foreach (token, key, val) {
		if (strequal(key, "name")) {
			strncpy(per->name, json_object_get_string(val),
				sizeof(per->name) - 1);
		} else if (strequal(key, "age")) {
			per->age = json_object_get_int(val);
		} else if (strequal(key, "blog")) {
			const char *blog = json_object_get_string(val);
			strncpy(per->blog, blog, sizeof(per->blog) - 1);
		} else if (strequal(key, "addr")) {
			const char *addr = json_object_get_string(val);
			strncpy(per->addr, addr, sizeof(per->addr) - 1);
		} else {
			printf("key=%s will be ignored\n", key);
		}
	}
}

void test_json_c(void)
{
	const char *json =
		R"({"name": "大大", "age": 30, "blog": "http://www.kkk.net","addr": "4414 spdd bbb"})";
	struct person p = {};

	do_test_json_c(json, &p);

	person_pretty_print(&p);
}

void test_any_struct(void)
{
	static struct {
		int age;
		char name[];
	} liulei __attribute__((aligned(8))) = {
		.name = "liulei",
		.age  = 38,
	};
	any_t v = ANY(liulei);
	(void)v;

	printf("sizeof(p)=%zd,alignof=%zd\n", sizeof(liulei), alignof(liulei));
}

int main(void)
{
	printf("__STDC_VERSION__=%ld\n", __STDC_VERSION__);
	test_json_c();
	return 0;
}
