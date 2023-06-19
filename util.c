#include <sys/socket.h>
#include <sys/types.h>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <netdb.h>

#include <stdio.h>
#include <limits.h>
#include <string.h>
#include <stdlib.h>
#include <stdbool.h>
#include <stdarg.h>

#include "util.h"

void fatal(const char *format, ...)
{
	va_list ap;

	va_start(ap, format);
	fprintf(stderr, "fatal error: ");
	vfprintf(stderr, format, ap);
	fprintf(stderr, "\n");
	va_end(ap);
	exit(EXIT_FAILURE);
}

static void Connect(int sockfd, struct sockaddr *addr)
{
	if (connect(sockfd, addr, sizeof(struct sockaddr_in)) < 0) {
		fatal("connect error: %m");
	}
}

struct keepalive_opts {
	int idle_secs;
	int probe_count;
	int probe_interval;
};

static inline void set_tcp_keepalive(int sockfd, struct keepalive_opts *options)
{
	int keepalive = 1;
	if (setsockopt(sockfd, SOL_SOCKET, SO_KEEPALIVE, &keepalive,
		       sizeof(keepalive)) < 0) {
		fatal("setsockopt error: %m");
	}
	setsockopt(sockfd, SOL_TCP, TCP_KEEPIDLE, &options->idle_secs,
		   sizeof(options->idle_secs));
	setsockopt(sockfd, SOL_TCP, TCP_KEEPINTVL, &options->probe_interval,
		   sizeof(options->probe_interval));
	setsockopt(sockfd, SOL_TCP, TCP_KEEPCNT, &options->probe_count,
		   sizeof(options->probe_count));
}

static inline int Getsockopt(int sockfd, int socklevel, int sockopt)
{
	int optval = 0;
	socklen_t len = sizeof(optval);
	if (getsockopt(sockfd, socklevel, sockopt, &optval, &len) < 0) {
		fatal("getsockopt error: %m");
	}
	return optval;
}

static void inspect_socket_options(int sockfd)
{
	bool keepalive;
	int probe_interval;
	int probe_count;
	int idle_seconds;

	keepalive = Getsockopt(sockfd, SOL_SOCKET, SO_KEEPALIVE);
	probe_interval = Getsockopt(sockfd, SOL_TCP, TCP_KEEPINTVL);
	probe_count = Getsockopt(sockfd, SOL_TCP, TCP_KEEPCNT);
	idle_seconds = Getsockopt(sockfd, SOL_TCP, TCP_KEEPIDLE);

	printf("tcp_keepalive=%s,idle_seconds=%d,probe_interval=%d,count=%d\n",
	       keepalive ? "on" : "off", idle_seconds, probe_interval,
	       probe_count);
}

static void inspect_addrinfo(const struct addrinfo *info)
{
	uint16_t port;
	struct sockaddr_in *addr;
	char ipv4[16];

	addr = (struct sockaddr_in *)info->ai_addr;
	port = ntohs(addr->sin_port);
	inet_ntop(addr->sin_family, &addr->sin_addr, ipv4, sizeof(ipv4));
	printf("addrinfo: famaily=%d socketype=%d protocol=%d ipv4=%s port=%hu\n",
	       info->ai_addr->sa_family, info->ai_socktype, info->ai_protocol,
	       ipv4, port);
}

void test_socket(void)
{
	int sockfd;
	struct addrinfo *info;
	struct addrinfo hints;

	memset(&hints, 0, sizeof(hints));
	hints.ai_family = AF_INET;
	hints.ai_socktype = SOCK_STREAM;

	if (getaddrinfo("www.baidu.com", "80", &hints, &info) < 0) {
		fatal("getaddrinfo error: %m");
	}
	for (; info != NULL; info = info->ai_next) {
		inspect_addrinfo(info);

		sockfd = socket(AF_INET, SOCK_STREAM, 0);

		Connect(sockfd, info->ai_addr);

		static struct keepalive_opts options = {
			.idle_secs = 9,
			.probe_count = 9,
			.probe_interval = 2,
		};

		set_tcp_keepalive(sockfd, &options);

		inspect_socket_options(sockfd);

		close(sockfd);
	}
}

union Optional {
	void *value;
	void *none;
};

void test_for_each(void)
{
	int list[] = { 1, 2, 3, 4, 5, 6, 7, 8 };
	int sum;

	sum = 0;
	for_each(list, lambda({ sum += x; }, int x));
	printf("sum = %d\n", sum);
}

struct options {
	char name[NAME_MAX];
	int flags;
};

#define Writeto(buf, size, ...) _vwrite(buf, size, ##__VA_ARGS__, NULL)

void _vwrite(void *buf, size_t size, ...)
{
	va_list ap;
	struct options *options;

	va_start(ap, size);
	for (;;) {
		options = va_arg(ap, struct options *);
		if (options == NULL) {
			break;
		}
		printf("options: id=%d, name=%s\n", options->flags,
		       options->name);
	}
	va_end(ap);
}

void test_writeto(void)
{
	struct options options;

	options = (struct options){ .name = "what", .flags = 123 };
	Writeto("123", 3, &options);
}
