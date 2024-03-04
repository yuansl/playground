#define _GNU_SOURCE
#include <sched.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <netdb.h> /* for getaddrinfo */
#include <complex.h>
#include <time.h>
#include <limits.h>
#include <math.h>
#include <stdio.h>
#include <stdatomic.h>
#include <stdint.h>
#include <unistd.h> /* for close */
#include <stdint.h>
#include <inttypes.h>
#include <pthread.h>
#include <string.h>

#include <jansson.h>

#include "util.h"
#include "stringbuffer.h"
#include "any.h"

#define STRING_BUFSIZE 10

#include "option.h"

#define IPV4_DOT_FORM_MAX_LEN 16

#define __noreturn	      __attribute__((noreturn))
#define __unused	      __attribute__((unused))

static inline void __nonnull((1)) inspect_sockaddr(struct sockaddr_in *inaddr)
{
	char ipbuf[IPV4_DOT_FORM_MAX_LEN];

	inet_ntop(inaddr->sin_family, &inaddr->sin_addr, ipbuf, sizeof(ipbuf));

	printf("connect to %s:%d\n", ipbuf, ntohs(inaddr->sin_port));
}

static int __nonnull((2))
	do_connect(int sockfd, const char *domain, const char *service)
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

#define NR_FLOATS 2

static void send_http_request(int connfd, const char host[static 1])
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

static void __unused test_http_request(void)
{
	const char *host = "www.qiniu.com";
	int connfd;

	connfd = initialize_connection(host, "http");
	if (connfd < 0)
		fatal("initialize_socket: %m\n");

	send_http_request(connfd, host);

	close(connfd);
}

typedef unsigned char byte;

struct sdshdr8 {
	uint8_t len;   /* used */
	uint8_t alloc; /* excluding the header and null terminator */
	byte flags;    /* 3 lsb of type, 5 unused bits */
	char buf[];
} __attribute__((__packed__));

typedef struct sdshdr8 sdshdr_t;

void test_sdshdr(void)
{
	sdshdr_t *shdr;
	byte *s;

	shdr = malloc(sizeof(sdshdr_t) + 1);
	s    = (byte *)shdr + sizeof(sdshdr_t);
	s[0] = '\0';
	free(shdr);
}

#define bitsizeof(x) (CHAR_BIT * sizeof(x))

#define maximum_unsigned_value_of_type(a) \
	(UINTMAX_MAX >> (bitsizeof(uintmax_t) - bitsizeof(a)))

#define unsigned_add_overflows(a, b) \
	((b) > maximum_unsigned_value_of_type(a) - (a))

#define die(...) fatal(__VA_ARGS__)

#define xcalloc	 calloc

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

	*opt = (struct option){
		.apply = options_set_age,
		.data  = malloc(sizeof(age)),
	};

	*(int *)opt->data = age;

	return opt;
}

option_t with_name(const char *name)
{
	option_t opt = malloc(sizeof(struct option));

	opt->apply = options_set_name;
	opt->data  = malloc(strlen(name) + 1);
	strcpy((char *)opt->data, name);

	return opt;
}

void init_options_internal(struct options *options, /* option_t opts */...)
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

#define init_options(opts, ...) init_options_internal(opts, __VA_ARGS__)

void with_retry(int retry_max, int (*f)(void))
{
	int attempts = 0;

	do {
		if (f() == 0)
			break;
		attempts++;
	} while (attempts <= retry_max);
}

static void __unused test_options(void)
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

void test_pthreads(void)
{
	pthread_t threads[NR_THREADS];

	pthread_create(&threads[0], NULL, print_some, NULL);
	pthread_create(&threads[1], NULL, inc_some, NULL);

	for (int i = 0; i < NR_THREADS; i++) {
		pthread_join(threads[i], NULL);
	}
}

struct point {
	double x, y;
};

union image {
	double pixel;
};

struct rectangle {
	struct point center;
	double width, height;
};

struct circle {
	struct point center;
	double r;
};

__nonnull((1)) static void scale_rectange_1p(struct rectangle *rec,
					     double scale)
{
	rec->width *= scale;
	rec->height *= scale;
}

static void __nonnull((1))
	scale_rectange_2p(struct rectangle *rec, double h_scale, double w_scale)
{
	rec->width *= w_scale;
	rec->height *= h_scale;
}

static void __nonnull((1)) scale_circle(struct circle *c, double scale)
{
	c->r *= scale;
}

/* clang-format off */

#define __scale2p(obj, ...)				\
	_Generic(obj,					\
		 struct rectangle*:scale_rectange_2p	\
                )(obj, __VA_ARGS__)

#define __scale1p(obj, ...)				\
	_Generic(obj,					\
	         struct circle*:scale_circle,		\
		 struct rectangle*:scale_rectange_1p	\
                )(obj, __VA_ARGS__)

#define __INVOKE_SCALE(_1,_2,_3,NAME,...) NAME 
	
#define scale(...) __INVOKE_SCALE(__VA_ARGS__, __scale2p, __scale1p)(__VA_ARGS__)

/* clang-format on */

struct array2 {
	int size;
	int cap;
	char buf[];
};

#define IMAGE(NAME, ...)    \
	(union image)       \
	{                   \
		__VA_ARGS__ \
	}

constexpr int NR_NUMBERS = 10;

void test_scale(void)
{
	union image img = IMAGE(.pixel = 3);
	(void)img;
	long double *numbers = calloc(NR_NUMBERS, sizeof(*numbers));
	struct array2 x	     = (static struct array2){ .buf = "hello, world" };

	printf("sizeof(double[12]) = %zd\n", sizeof(double[12]));
	(void)x;

	for (int i = 0; i < NR_NUMBERS; i++)
		numbers[i] = rand() % 100;

	struct rectangle r = { .height = 1, .width = 2 };
	struct circle c	   = { .r = 5 };

	scale(&r, 3.8, 3.5);
	scale(&c, 4.2);

	printf("rectange.height=%.2f,wigth=%.2f, circle.r=%.2f\n", r.height,
	       r.width, c.r);
}

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

void __unused test_any(void)
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

	return a;
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

#ifndef NAME_MAX
#define NAME_MAX 256
#endif

enum gender : uint8_t {
	MALE,
	FEMALE
};

struct person {
	char name[NAME_MAX];
	int age;
	char blog[NAME_MAX];
	char addr[NAME_MAX];
	enum gender gender;
};

const struct person __unused *liuhai = NULL;

static void person_pretty_print(struct person *p)
{
	printf("name: %s, age: %d, blog: '%s', addr: '%s'\n", p->name, p->age,
	       p->blog, p->addr);
}

void do_test_jansson(const char json[static 1], struct person *__unused)
{
	json_t *object;
	const char *key;
	json_t *val;
	json_error_t err;

	object = json_loadb(json, strlen(json), JSON_DECODE_ANY, &err);
	if (!object)
		fatal("json_loads: %m");

	printf("load json successfully\n");

	switch (json_typeof(object)) {
	case JSON_INTEGER:
		json_integer_value(object);
		break;
	case JSON_STRING:
		break;
	case JSON_OBJECT:
		printf("this is a json_object\n");

	default:
		break;
	}

	json_object_foreach (object, key, val) {
		/*
		 * if (strequal(key, "name")) {
		 * 	const char *name = json_string_value(val);
		 * 	printf("name=%s\n", name);
		 * 	strncpy(per->name, name, sizeof(per->name) - 1);
		 * } else if (strequal(key, "age")) {
		 * 	per->age = json_integer_value(val);
		 * } else if (strequal(key, "blog")) {
		 * 	const char *blog = json_string_value(val);
		 * 	strncpy(per->blog, blog, sizeof(per->blog) - 1);
		 * } else if (strequal(key, "addr")) {
		 * 	const char *addr = json_string_value(val);
		 * 	strncpy(per->addr, addr, sizeof(per->addr) - 1);
		 * } else {
		 * 	printf("key=%s will be ignored\n", key);
		 * }
		 */
	}
}

#define array_of(name, type, cap)                             \
	struct array_##type {                                 \
		size_t size;                                  \
		type data[];                                  \
	};                                                    \
	struct array_##type *name =                           \
		malloc(sizeof(*(name)) + cap * sizeof(type)); \
	name->size = cap

#define RAW(s) R"()"
/*
 * R"({"name": "大大", "age": 30, "blog": "http://www.kkk.net", "addr": "4414
 * spdd bbb"})"
 */
void test_json_parser(void)
{
	const char *json = "";
	struct person p	 = {};

	do_test_jansson(json, &p);

	person_pretty_print(&p);
}

void test_any_struct(void)
{
	static struct {
		int age;
		char name[];
	} liulei = {
		.name = "liulei",
		.age  = 38,
	};
	any_t v = ANY(liulei);
	(void)v;

	printf("sizeof(p)=%zd,alignof=%zd\n", sizeof(liulei), alignof(liulei));
}

struct shared {
	_Atomic int count;
	_Atomic bool done;
};

struct thread {
	struct shared *shared;
	pthread_t id;
};

void *count_self(void *arg)
{
	struct thread *thread = arg;

	nanosleep(&(struct timespec){ .tv_nsec = rand() % 1000 }, NULL);

	thread->shared->count++;

	if (pthread_self() % 2 == 0) {
		printf("in thread %ld, count now = %d\n",
		       pthread_self() % NR_THREADS,
		       atomic_load(&thread->shared->count));

		thread->shared->done = true;
	}

	pthread_exit(NULL);
}

void test_atomic_variable(void)
{
	struct shared shared		  = {};
	struct thread threads[NR_THREADS] = {
		[0 ... NR_THREADS - 1].shared = &shared,
	};

	for (int i = 0; i < NR_THREADS; i++) {
		int e;
		threads[i].shared = &shared;
		e = pthread_create(&threads[i].id, NULL, count_self,
				   &threads[i]);
		if (e)
			fatal("pthread_create: %s\n", strerror(e));
	}

	for (int i = 0; i < NR_THREADS; i++) {
		pthread_join(threads[i].id, NULL);
	}

	if (shared.count != NR_THREADS)
		printf("Oops: shared.count=%d, expected = %d\n", shared.count,
		       NR_THREADS);
}

struct refcount {
	_Atomic int count;
};

#define REFCOUNT_DEC(ref)  --(ref).count
#define REFCOUNT_INCR(ref) (ref).count++

struct some {
	char name[256];
	int age;
	enum gender gender;
	struct refcount refcount;
};

struct some *clone_some(struct some *some)
{
	REFCOUNT_INCR(some->refcount);
	return some;
}

static struct some *some_init(struct some *some)
{
	*some = (struct some){
		.age		= 26,
		.name		= "hello",
		.gender		= MALE,
		.refcount.count = 1,
	};
	return some;
}

void cleanup_x(void *x)
{
	struct some *some = x;

	if (--some->refcount.count <= 0) {
		printf("now , variable x(struct some*) will be freed");
		free(some);
	}

	printf("hello, this is a cleanup function %d\n", *(int *)x);
}

#define lock(x) atomic_compare_exchange_strong(&x, &(int){ 0 }, 1)

void unlock(void *x)
{
	printf("unlock (x)...\n");
	atomic_compare_exchange_strong((_Atomic int *)x, &(int){ 1 }, 0);
}

double chainable_division(double x, double y)
{
	return x / y;
}

#define __native_word(t)                                            \
	(sizeof(t) == sizeof(char) || sizeof(t) == sizeof(short) || \
	 sizeof(t) == sizeof(int) || sizeof(t) == sizeof(long))

enum kind {
	KIND_ENUM	      = 1,
	KIND_POINTER_OR_ARRAY = 5,
	KIND_DECIMAL	      = 8,
	KIND_COMPLEX	      = 9,
	KIND_STRUCT	      = 12,
	KIND_UNION	      = 13
};

void test_classify_type(void)
{
	struct V {
		char buf1[10];
		int b;
		char buf2[10];
	} var;
	void *p = &var.buf1[1], *q = &var.b;

	(void)p;

	bool k __unused	      = true;
	enum kind k2 __unused = KIND_ENUM;
	int k3 __unused	      = 0;
	long k4 __unused      = 0L;
	char k5 __unused      = 'a';
	typedef void(func)(void);
	func *k11 __unused;
	union {
		int a;
		double b;
		long c;
		char d[256];
	} k12	    = {};
	int k13[12] = { [0 ... 9] = 1 };
	char *k14   = "hello";

	(void)k12;
	(void)k13;

	printf("address q=%p,&var=%p, %zd, typeof(struct V)=%d\n", q, &var,
	       (char *)(&var + 1) - (char *)q, __builtin_classify_type(k14));
}

static void some_destory(void *some)
{
	struct some *_some = *(struct some **)some;

	fprintf(stderr, "hello, some\n");

	if (REFCOUNT_DEC(_some->refcount) <= 0) {
		printf("OK, some will be freed\n");
		free(_some);
	}
}

void test_cleanup(void)
{
	struct some *b;
	{
		struct some *some __attribute__((cleanup(some_destory)));
		some = malloc(sizeof(*some));

		some_init(some);

		b = clone_some(some);

		printf("some={.age=%d,.name='%s'.refcount=%d}\n", some->age,
		       some->name, some->refcount.count);
	}
	size_t b_size = __builtin_object_size(b, 0);
	int type      = __builtin_classify_type(b);
	printf("b_size=%zd, sizeof *b=%zd, b={.age=%d,.name='%s',refcount=%d},native_type(*b)=%s, alignof(b)=%zd, type(b)=%d\n",
	       b_size, sizeof(*b), b->age, b->name, b->refcount.count,
	       __native_word(*b) ? "true" : "false", alignof(struct some *),
	       type);

	{
		_Atomic int x __attribute__((cleanup(unlock))) = 0;
		if (lock(x)) {
			printf("lock acuired\n");
		}
	}
	if (isinf(4 / 0.0))
		printf("4/0 is inf\n");

	double x = chainable_division(2.0, chainable_division(4, 0.0));
	char *px = (char *)&x;
	printf("x = %.2f\n", x);

	for (int i = 0; i < (int)sizeof(x); i++) {
		px[i] = 0xff;
	}

	if (isnan(x))
		printf("now x is Nan\n");
}

#define pretty_print(format, ...)                               \
	do {                                                    \
		time_t now = time(NULL);                        \
		char line[LINE_MAX];                            \
		int n	= strftime(line, sizeof(line), "%T %F", \
				   localtime(&now));            \
		line[n] = '\0';                                 \
		n += snprintf(line + n, sizeof(line) - n - 1,   \
			      ":%s:%d: ", __FILE__, __LINE__);  \
		snprintf(line + n, sizeof(line) - n - 1,        \
			 format __VA_OPT__(, ) __VA_ARGS__);    \
		fprintf(stderr, "%s\n", line);                  \
	} while (0)

#define print(format, ...)                                       \
	do {                                                     \
		pretty_print(format __VA_OPT__(, ) __VA_ARGS__); \
	} while (0)

int ab = 3;

union tuple {
	int _1;
	long _2;
	double _3;
};

static int change(int *b, union tuple *c)
{
	(void)b;
	c->_2 = 30;

	return ab;
}

int main(void)
{
	int cpu;

	cpu = sysconf(_SC_NPROCESSORS_ONLN);
	print("cpu number");
	pretty_print("The thread is running on cpu #%1$u, %1$u, %1$u\n", cpu);

	{
		change_a(&ab, &ab);

		printf("ab = %d\n", ab);
	}

#define TEST_CASE()
	TEST_CASE()
	{
		printf("-1=%0#b\n", -2 << 12);
	}

	return 0;
}
