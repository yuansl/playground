/* copied from https://www.kernel.org/doc/html/v6.5/input/uinput.html */
#include <linux/uinput.h>

#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <fcntl.h>
#include <string.h>

__attribute__((noreturn)) static inline void __fatal(const char *fmt, ...)
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
		__fatal(__VA_ARGS__);                                        \
	})

static void emit(int fd, int type, int code, int val)
	__attribute__((fd_arg(1)));

static void emit(int fd, int type, int code, int val)
{
	struct input_event ie = {};

	ie.type = type;
	ie.code = code;
	ie.value = val;
	/* timestamp values below are ignored */
	ie.time.tv_sec = 0;
	ie.time.tv_usec = 0;

	(void)write(fd, &ie, sizeof(ie));
}

#define UINPUT_EVENT_DETECT_TIME_SECONDS 80

int main(void)
{
	struct uinput_setup usetup = {};
	int fd;

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

	usetup.id.bustype = BUS_USB;
	usetup.id.vendor = 0x1234;  /* sample vendor */
	usetup.id.product = 0x5678; /* sample product */
	strcpy(usetup.name, "Example device");

	ioctl(fd, UI_DEV_SETUP, &usetup);
	ioctl(fd, UI_DEV_CREATE);

	/*
	 * On UI_DEV_CREATE the kernel will create the device node for this
	 * device. We are inserting a pause here so that userspace has time
	 * to detect, initialize the new device, and can start listening to
	 * the event, otherwise it will not notice the event we are about
	 * to send. This pause is only needed in our example code!
	 */

	usleep(UINPUT_EVENT_DETECT_TIME_SECONDS * 1000);

	/* Key press, report the event, send key release, and report again */
	emit(fd, EV_KEY, KEY_LEFTMETA, 1);
	emit(fd, EV_SYN, SYN_REPORT, 0);
	emit(fd, EV_KEY, KEY_E, 1);
	emit(fd, EV_SYN, SYN_REPORT, 0);
	emit(fd, EV_KEY, KEY_E, 0);
	emit(fd, EV_SYN, SYN_REPORT, 0);
	emit(fd, EV_KEY, KEY_LEFTMETA, 0);
	emit(fd, EV_SYN, SYN_REPORT, 0);

	/*
	 * Give userspace some time to read the events before we destroy the
	 * device with UI_DEV_DESTOY.
	 */
	usleep(UINPUT_EVENT_DETECT_TIME_SECONDS * 1000);

	ioctl(fd, UI_DEV_DESTROY);
	close(fd);

	return 0;
}
