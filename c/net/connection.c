#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <stdint.h>
#include <stdlib.h>
struct connection;

enum connection_status { RUNNING, IDLE, WAITNG, CLOSED };

struct refcount {
	_Atomic int count;
};

struct tcp_connection {
	int sockfd;
	enum connection_status status;
	int ttl;
	struct refcount refcount;
};

struct udp_connection;

#define ARRAY_SIZE(a) (sizeof(a) / sizeof(a[0]))

struct tcp_connection *available_connections[] = {};

struct tcp_connection *pick_available_connection(const char *ip, uint16_t port)
{
	for (int i = 0; i < (int)ARRAY_SIZE(available_connections); i++) {
		struct tcp_connection *conn = available_connections[i];
		if (conn && conn->status == IDLE) {
			return conn;
		}
	}
	return NULL;
}

static void init_sockaddr_inet(struct sockaddr_in *addr, const char *ip,
			       in_port_t port)
{
	in_addr_t a;

	inet_pton(AF_INET, ip, &a);
	*addr = (struct sockaddr_in){
		.sin_family	 = AF_INET,
		.sin_addr.s_addr = a,
		.sin_port	 = htons(port),
	};
}

int tcp_connect(const char *ip, uint16_t port)
{
	int sockfd;
	struct sockaddr_in addr;

	sockfd = socket(AF_INET, SOCK_STREAM, 0);
	if (sockfd < 0) {
		return -1;
	}

	init_sockaddr_inet(&addr, ip, port);
	if (connect(sockfd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
		return -1;
	}
	return sockfd;
}

#define CONNECTION_TTL_DEFAULT 20

struct tcp_connection *create_tcp4_connection(const char *ip, uint16_t port)
{
	struct tcp_connection *conn;
	int sockfd;

	if ((sockfd = tcp_connect(ip, port)) < 0) {
		return NULL;
	}

	conn  = malloc(sizeof(*conn));
	*conn = (struct tcp_connection){
		.sockfd		= sockfd,
		.status		= IDLE,
		.ttl		= CONNECTION_TTL_DEFAULT,
		.refcount.count = 1,
	};
	return conn;
}

struct tcp_connection *create_tcp6_connection(const char *ip6, uint16_t port);

struct tcp_connection *get_connection(const char *ip, uint16_t port)
{
	struct tcp_connection *conn;

	conn = pick_available_connection(ip, port);
	if (!conn)
		conn = create_tcp4_connection(ip, port);
	return conn;
}
