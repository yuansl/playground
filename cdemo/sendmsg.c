#include <sys/socket.h>
#include <sys/types.h>
#include <sys/un.h>

#include <unistd.h>
#include <fcntl.h>
#include <string.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>

#define NUM_FDS 1
#define FD_SOCKET "/tmp/someserver"

static inline void fatal(const char *fmt, ...)
{
	va_list ap;

	va_start(ap, fmt);
	fprintf(stderr, "fatal error: ");
	vfprintf(stderr, fmt, ap);
	va_end(ap);

	exit(1);
}

int main(int argc, char *argv[])
{
	int sockfd;
	int fds[NUM_FDS] = { 0 };
	union {
		char buf[CMSG_SPACE(sizeof(fds))];
		struct cmsghdr __padding;
	} u;
	struct msghdr msg = { 0 };
	struct cmsghdr *cmsg;
	struct sockaddr_un addr;

	sockfd = socket(AF_UNIX, SOCK_DGRAM, 0);

	addr = (struct sockaddr_un) {
		.sun_path = FD_SOCKET,
		.sun_family = AF_UNIX
	};
	msg.msg_name = &addr;
	msg.msg_namelen = sizeof(addr);
	msg.msg_control = u.buf;
	msg.msg_controllen = sizeof(u.buf);

	if ((fds[0] = open("/tmp/some2", O_RDWR | O_CREAT | O_TRUNC, 0666)) < 0) {
		fatal("open error: %m\n");
	}

	cmsg = CMSG_FIRSTHDR(&msg);
	cmsg->cmsg_level = SOL_SOCKET;
	cmsg->cmsg_type = SCM_RIGHTS;
	cmsg->cmsg_len = CMSG_LEN(sizeof(fds));
	*(int *)CMSG_DATA(cmsg) = fds[0];

	if (sendmsg(sockfd, &msg, 0) < 0) {
		fatal("sendmsg: %m\n");
	}
	char buf[BUFSIZ];
	ssize_t n;

	printf("read offset of fd %d: %zd\n", fds[0], lseek(fds[0], 0, SEEK_CUR));

	lseek(fds[0], 0, SEEK_SET);
	
	if ((n=read(fds[0], buf, sizeof(buf))) < 0) {
		fatal("read: %m\n");
	}
	buf[n] = '\0';

	printf("Read message: `%s`\n", buf);

	return 0;
}
