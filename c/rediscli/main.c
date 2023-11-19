#define _GNU_SOURCE
#include <unistd.h>
#include <stdio.h>
#include <limits.h>

struct redis {
	/* TODO */
};

typedef struct redis *redis_t;

struct redis_commands {
	void (*get)(redis_t, const char *key);
	void (*set)(redis_t, const char *key, const char *value);
	void (*hset)(redis_t, const char *key, const char *field1,
		     const char *value1, ...);
	void (*hget)(redis_t, const char *key, const char *field1,
		     const char *field2, ...);
	void (*zrevrange)();
	void (*zcard)();
	void (*scard)();
	void (*smembers)();
	void (*bpop)();
	void (*bpush)();
};

struct command_result {
	int flags;
};

void pretty_environ(void)
{
	for (int i = 0; environ[i] != NULL; i++) {
		printf("environment: `%s`\n", environ[i]);
	}
}

int main(int argc, char *argv[])
{
	/*
	 * redis_t *rdb = redis_connect("redis://mint.local:6379,x1carbon.local:6379:darwin.local:6379/0?write_timeout=2s")
	 * redis_set(rdb, "some", 1);
	 * 
	 * int_cmd *result = redis_get(rdb, "some");
	 * assert(result->value == 1);
	 * 
	 * redis_hset(rdb, "defy:fscdn:domaindict:google", "www.example.com", bytes);
	 *
	 * string_cmd *result;
	 * result = redis_hget(rdb, "defy:fscdn:domaindict:google", "www.example.com");
	 * if (!result) {
	 * 	return 1;
	 * }
	 */
	return 0;
}
