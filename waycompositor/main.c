#define _GNU_SOURCE
#include <sys/socket.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>

#include <wayland-client.h>

#define PROCESS_NAME_MAX_LENGTH 1024

static char *read_from(const char *filename, char buf[], size_t bufsz)
{
	FILE *fp;

	fp = fopen(filename, "r");
	if (!fp) {
		fprintf(stderr, "opening '%s' failed: %s\n", filename,
			strerror(errno));
		return NULL;
	}
	if (!fgets(buf, bufsz, fp)) {
		fprintf(stderr, "reading '%s' failed\n", filename);
		fclose(fp);
		return NULL;
	}
	buf[strcspn(buf, "\n")] = '\0';
	fclose(fp);

	return buf;
}


static pid_t pid_from_fd(int fd)
{
	struct ucred ucred;
	socklen_t len = sizeof(struct ucred);

	if (getsockopt(fd, SOL_SOCKET, SO_PEERCRED, &ucred, &len) == -1) {
		perror("getsockopt failed");
		exit(-1);
	}
	return ucred.pid;
}

static char *get_process_name_of(pid_t pid, char name[], size_t size)
{
	char proc_buf[64];
		
	sprintf(proc_buf, "/proc/%d/comm", pid);

	return read_from(proc_buf, name, size);
}

void show_process_info_of(struct wl_display *display)
{
	char process_name[PROCESS_NAME_MAX_LENGTH];
	pid_t pid;
	
	pid = pid_from_fd(wl_display_get_fd(display));

	if (!get_process_name_of(pid, process_name, sizeof(process_name))) {
		fprintf(stderr, "process_name_from_pid error:%m\n");
		exit(EXIT_FAILURE);
	}
	printf("pid: %d, process name: %s\n", pid, process_name);
}

int main(void)
{
	struct wl_display *display;

	display = wl_display_connect(NULL);
	if (!display) {
		fprintf(stderr, "can't connect to display\n");
		return 1;
	}

	show_process_info_of(display);

	wl_display_disconnect(display);

	return 0;
}
