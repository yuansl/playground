/* copied from https://www.kernel.org/doc/html/v6.5/input/uinput.html */
#include <linux/uinput.h>
#include <unistd.h>
#include <fcntl.h>

#include <signal.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

[[noreturn]] static inline void _fatal(const char *fmt, ...)
{
	va_list ap;

	va_start(ap);
	vfprintf(stderr, fmt, ap);
	fprintf(stderr, "\n");
	va_end(ap);
	exit(EXIT_FAILURE);
}

#define fatal(...)                                                           \
	({                                                                   \
		fprintf(stderr, "%s:%d: fatal error: ", __func__, __LINE__); \
		_fatal(__VA_ARGS__);                                        \
	})

static void emit_input_event(int fd, int type, int code, int val)
	__attribute__((fd_arg(1)));

static void emit_input_event(int fd, int type, int code, int val)
{
	struct input_event ie = {};

	ie.type = type;
	ie.code = code;
	ie.value = val;
	/* timestamp values below are ignored */
	ie.time.tv_sec = 0;
	ie.time.tv_usec = 0;

	if (write(fd, &ie, sizeof(ie)) != sizeof(ie))
		fatal("write(fd,&ie): %m");
}

#define UINPUT_EVENT_DETECT_TIME_SECONDS 80

static inline int create_uinput_dev_keyboard(void)
{
	int fd;
	struct uinput_setup usetup = {};

	fd = open("/dev/uinput", O_WRONLY | O_NONBLOCK);
	if (fd < 0) {
		fatal("open(/dev/uinput): %m");
	}
	/*
	 * The ioctls below will enable the device that is about to be
	 * created, to pass key events, in this case the Super(Meta)-E key.
	 */
	ioctl(fd, UI_SET_EVBIT, EV_KEY);
	ioctl(fd, UI_SET_KEYBIT, KEY_LEFTMETA);
	ioctl(fd, UI_SET_KEYBIT, KEY_E);
	usetup = (struct uinput_setup){
		.id = (struct input_id){ .bustype = BUS_USB,
					 .vendor = 0x1,
					 .product = 0x1 },
		.name = "uinput keyboard"
	};
	ioctl(fd, UI_DEV_SETUP, &usetup);
	ioctl(fd, UI_DEV_CREATE);

	/*
	 * On UI_DEV_CREATE the kernel will create the device node for this
	 * device. We are inserting a pause here so that userspace has time
	 * to detect, initialize the new device, and can start listening to
	 * the event, otherwise it will not notice the event we are about
	 * to send. This pause is only needed in our example code!
	 */
	/*
	 * usleep(UINPUT_EVENT_DETECT_TIME_SECONDS * 1000);
	 */

	return fd;
}

static inline void destory_uinput_dev_keyboard(int fd)
{
	ioctl(fd, UI_DEV_DESTROY);
	close(fd);
}

static inline void press_leftmeta_e_key_and_release(int fd)
{
	/* Key press, report the event, send key release, and report again */
	emit_input_event(fd, EV_KEY, KEY_LEFTMETA, 1);
	emit_input_event(fd, EV_SYN, SYN_REPORT, 0);
	emit_input_event(fd, EV_KEY, KEY_E, 1);
	emit_input_event(fd, EV_SYN, SYN_REPORT, 0);
	emit_input_event(fd, EV_KEY, KEY_E, 0);
	emit_input_event(fd, EV_SYN, SYN_REPORT, 0);
	emit_input_event(fd, EV_KEY, KEY_LEFTMETA, 0);
	emit_input_event(fd, EV_SYN, SYN_REPORT, 0);
}


int pidfd;

const char *pid_file_name = "/run/raise-emacs.pid";

static void delete_pid_file(void)
{
	unlink(pid_file_name);
}

static void save_pid_to_file(void)
{
	char buf[BUFSIZ];

	pidfd = open(pid_file_name, O_CREAT | O_EXCL | O_WRONLY, 0660);
	if (pidfd < 0)
		fatal("open('%s'): %m\n", pid_file_name);

        sprintf(buf, "%d", getpid());

	if (write(pidfd, buf, strlen(buf)) != (ssize_t)strlen(buf))
		fatal("write(pidfd): %m\n");
}

int fd;

static void sigusr1_handler(int)
{
	press_leftmeta_e_key_and_release(fd);
}

[[noreturn]] static void sigterm_handler(int)
{
	delete_pid_file();
	exit(EXIT_FAILURE);
}

int main(void)
{
	fd = create_uinput_dev_keyboard();

	if (daemon(0, 0) < 0)
		fatal("daemon: %m\n");

	save_pid_to_file();
	atexit(delete_pid_file);

	signal(SIGUSR1, sigusr1_handler);
	signal(SIGTERM, sigterm_handler);

	while (true) {
		pause();
	}

	/*
	 * Give userspace some time to read the events before we destroy the
	 * device with UI_DEV_DESTOY.
	 */
	/*
	 * usleep(UINPUT_EVENT_DETECT_TIME_SECONDS * 1000);
	 */

	destory_uinput_dev_keyboard(fd);

	return 0;
}
