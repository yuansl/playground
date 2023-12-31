#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <threads.h>
#include <time.h>
#include <unistd.h>
#include <pthread.h>
#include <threads.h>

#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <stdarg.h>
#include <inttypes.h>
#include <stdint.h>
#include <limits.h>

#include "option.h"

#define IPV4_DOT_FORM_MAX_LEN 16

[[noreturn]] static inline void _fatal(const char *fmt, ...)
{
	va_list ap;

	va_start(ap);
	vfprintf(stderr, fmt, ap);
	va_end(ap);
	exit(EXIT_FAILURE);
}

#define fatal(...)                                                           \
	({                                                                   \
		fprintf(stderr, "%s:%d: fatal error: ", __FILE__, __LINE__); \
		_fatal(__VA_ARGS__);                                         \
	})

#define panic(...)                                                   \
	({                                                           \
		fprintf(stderr, "%s:%d: BUG: ", __FILE__, __LINE__); \
		abort();                                             \
	})

static inline void inspect_sockaddr(struct sockaddr_in *inaddr)
{
	char ipbuf[IPV4_DOT_FORM_MAX_LEN];

	inet_ntop(inaddr->sin_family, &inaddr->sin_addr, ipbuf, sizeof(ipbuf));

	printf("connect to %s:%d\n", ipbuf, ntohs(inaddr->sin_port));
}

static int do_connect(int sockfd, const char *domain, const char *service)
{
	struct addrinfo *addrs;
	bool found;
	int err;

	if ((err = getaddrinfo(domain, service, NULL, &addrs)) != 0) {
		fprintf(stderr, "getaddrinfo: %s\n", gai_strerror(err));
		return -1;
	}

	found = false;
	for (struct addrinfo *ai = addrs; ai != NULL; ai = ai->ai_next) {
		inspect_sockaddr((struct sockaddr_in *)ai->ai_addr);

		if (connect(sockfd, ai->ai_addr, ai->ai_addrlen) == 0) {
			found = true;
			break;
		}
	}
	freeaddrinfo(addrs);

	return found ? 0 : -1;
}

static int initialize_connection(const char *domain, const char *service)
{
	int sockfd;

	if ((sockfd = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
		perror("socket");
		return -1;
	}
	if (do_connect(sockfd, domain, service) < 0) {
		close(sockfd);
		return -1;
	}
	return sockfd;
}

static void send_http_request(int connfd, const char *host)
{
	char buf[BUFSIZ];
	ssize_t nbytes;

	sprintf(buf,
		"GET /index.html HTTP/1.1\r\nHost: %s\r\nUser-Agent: Simple-http/1.0\r\n\r\n",
		host);
	nbytes = strlen(buf);

	printf("request(raw): '%s'\n", buf);

	if (write(connfd, buf, nbytes) != nbytes)
		fatal("write: %m\n");

	if ((nbytes = read(connfd, buf, sizeof(buf))) < 0)
		fatal("read: %m\n");
	buf[nbytes] = '\0';

	printf("response(raw)='%s'\n", buf);
}

typedef struct {
} *json_t;
typedef json_t json_value;

json_t json_parse(const char *s);
json_value json_get(json_t, const char *key);
int json_put(json_t, const char *key, const json_value value);

[[maybe_unused]] static void test_http_request(void)
{
	const char *host = "www.qiniu.com";
	int connfd;

	connfd = initialize_connection(host, "http");
	if (connfd < 0)
		fatal("initialize_socket: %m\n");

	send_http_request(connfd, host);

	close(connfd);
}

struct sdshdr8 {
	uint8_t len;	     /* used */
	uint8_t alloc;	     /* excluding the header and null terminator */
	unsigned char flags; /* 3 lsb of type, 5 unused bits */
	char buf[];
} __attribute__((__packed__));

typedef struct sdshdr8 sdshdr_t;

void test_sdshdr(void)
{
	sdshdr_t *shdr;
	unsigned char *s;

	shdr = malloc(sizeof(sdshdr_t) + 1);
	s = (unsigned char *)shdr + sizeof(sdshdr_t);
	s[0] = '\0';
	free(shdr);
}

#define bitsizeof(x) (CHAR_BIT * sizeof(x))

#define maximum_unsigned_value_of_type(a) \
	(UINTMAX_MAX >> (bitsizeof(uintmax_t) - bitsizeof(a)))

#define unsigned_add_overflows(a, b) \
	((b) > maximum_unsigned_value_of_type(a) - (a))

#define die(...) fatal(__VA_ARGS__)

#define xcalloc calloc

static inline size_t st_add(size_t a, size_t b)
{
	if (unsigned_add_overflows(a, b))
		die("size_t overflow: %" PRIuMAX " + %" PRIuMAX, (uintmax_t)a,
		    (uintmax_t)b);
	return a + b;
}
#define st_add3(a, b, c) st_add(st_add((a), (b)), (c))

#define FLEX_ALLOC_MEM(x, flexname, buf, len)                                \
	do {                                                                 \
		size_t flex_array_len_ = (len);                              \
		(x) = xcalloc(1, st_add3(sizeof(*(x)), flex_array_len_, 1)); \
		memcpy((void *)(x)->flexname, (buf), flex_array_len_);       \
	} while (0)
#define FLEXPTR_ALLOC_MEM(x, ptrname, buf, len)                              \
	do {                                                                 \
		size_t flex_array_len_ = (len);                              \
		(x) = xcalloc(1, st_add3(sizeof(*(x)), flex_array_len_, 1)); \
		memcpy((x) + 1, (buf), flex_array_len_);                     \
		(x)->ptrname = (void *)((x) + 1);                            \
	} while (0)
#define FLEX_ALLOC_STR(x, flexname, str) \
	FLEX_ALLOC_MEM((x), flexname, (str), strlen(str))
#define FLEXPTR_ALLOC_STR(x, ptrname, str) \
	FLEXPTR_ALLOC_MEM((x), ptrname, (str), strlen(str))

struct array {
	size_t alloc;
	size_t nr;
	char name[];
};

void test_array(void)
{
	struct array *f;

	FLEX_ALLOC_STR(f, name, "something");

	free(f);
}

struct options {
	int age;
	const char *name;
};

static void options_set_age(option_t opt, void *options)
{
	((struct options *)options)->age = *(int *)opt->data;
}

static void options_set_name(option_t opt, void *options)
{
	((struct options *)options)->name = opt->data;

	/* Ok, we reuse opt->data here, avoid allocating memory again */
	opt->data = NULL;
}

option_t with_age(int age)
{
	option_t opt = malloc(sizeof(struct option));

	opt->apply = options_set_age;
	opt->data = malloc(sizeof(age));
	*(int *)opt->data = age;

	return opt;
}

option_t with_name(const char *name)
{
	option_t opt = malloc(sizeof(struct option));

	opt->apply = options_set_name;
	opt->data = malloc(strlen(name) + 1);
	strcpy((char *)opt->data, name);

	return opt;
}

void init_options(struct options *options, /* option_t opts */...)
{
	va_list ap;

	va_start(ap, options);
	while (true) {
		option_t opt = va_arg(ap, option_t);
		if (!opt)
			break;
		opt->apply(opt, options);

		option_destroy(opt);
	}
	va_end(ap);
}

void with_retry(int retry_max, int (*f)(void))
{
	int attempts = 0;

	do {
		if (f() == 0)
			break;
		attempts++;
	} while (attempts <= retry_max);
}

[[maybe_unused]] static void test_options(void)
{
	struct options options = {};

	init_options(&options, with_age(38), with_name("liming"), NULL);

	printf("age=%d, name='%s'\n", options.age, options.name);
}

thread_local int some_key;

void *print_some(void *)
{
	for (int i = 0; i < 10; i++) {
		printf("print thread %ld: thread_local key=%d\n",
		       pthread_self() % 1000'000, some_key);
		nanosleep(&(static struct timespec){ .tv_sec = 1 }, NULL);
	}
	pthread_exit(NULL);
}

void *inc_some(void *)
{
	for (int i = 0; i < 10; i++) {
		printf("inc thread %ld: thread_local key=%d\n",
		       pthread_self() % 1000'000, some_key);
		some_key++;
		nanosleep(&(static struct timespec){ .tv_nsec = 1000000 },
			  NULL);
	}
	pthread_exit(NULL);
}

#define NR_THREADS 2

int main(void)
{
	pthread_t threads[NR_THREADS];

	pthread_create(&threads[0], NULL, print_some, NULL);
	pthread_create(&threads[1], NULL, inc_some, NULL);

	for (int i = 0; i < NR_THREADS; i++) {
		pthread_join(threads[i], NULL);
	}
	return 0;
}
