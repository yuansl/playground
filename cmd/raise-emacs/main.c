#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>

#include <jansson.h>
#include <dbus/dbus.h>

#define __unused       __attribute__((unused))

#define DBUS_NAME      "org.gnome.Shell"
#define DBUS_PATH      "/org/gnome/Shell/Extensions/WindowsExt"
#define DBUS_INTERFACE "org.gnome.Shell.Extensions.WindowsExt"

struct dbus_context {
	DBusConnection *conn;
};

void dbus_init(struct dbus_context *ctx)
{
	DBusError err;

	// initialise the errors
	dbus_error_init(&err);

	// connect to the bus
	ctx->conn = dbus_bus_get_private(DBUS_BUS_SESSION, &err);

	if (dbus_error_is_set(&err)) {
		fprintf(stderr, "Connection Error (%s)\n", err.message);
		dbus_error_free(&err);
	}
	if (!ctx->conn) {
		exit(1);
	}
}

typedef DBusMessage dbus_message;

static dbus_message *dbus_call_method(struct dbus_context *ctx,
				      dbus_message *msg)
{
	DBusPendingCall *pending;

	// send message and get a handle for a reply
	if (!dbus_connection_send_with_reply(ctx->conn, msg, &pending,
					     -1)) { // -1 is default
						    // timeout
		fprintf(stderr, "Out Of Memory!\n");
		exit(1);
	}
	if (!pending) {
		fprintf(stderr, "Pending Call Null\n");
		exit(1);
	}
	dbus_connection_flush(ctx->conn);
	// free message
	dbus_message_unref(msg);
	// block until we receive a reply
	dbus_pending_call_block(pending);
	// get the reply message
	msg = dbus_pending_call_steal_reply(pending);
	if (!msg) {
		fprintf(stderr, "Reply Null\n");
		exit(1);
	}
	// free the pending message handle
	dbus_pending_call_unref(pending);

	return msg;
}

void dbus_close(struct dbus_context *ctx)
{
	dbus_connection_close(ctx->conn);
}

struct wl_window_desc {
	char *class;
	char *title;
	pid_t pid;
	uint64_t id;
	bool focus;
	int64_t maximized;
};

static void wl_window_pretty(struct wl_window_desc *window)
{
	printf("title=%s,pid=%d,id=%lu,focus=%s,maximized=%ld\n", window->title,
	       window->pid, window->id, window->focus ? "true" : "false",
	       window->maximized);
}

static void json_decode_wl_window(json_t *iter, void *v)
{
	json_t *class		      = json_object_get(iter, "class");
	struct wl_window_desc *window = v;

	assert(class != NULL && v != NULL);

	switch (class->type) {
	case JSON_STRING:
		if (!strcmp(json_string_value(class), "emacs")) {
			window->title = (char *)json_string_value(
				json_object_get(iter, "title"));
			window->id =
				json_integer_value(json_object_get(iter, "id"));
			window->pid = json_integer_value(
				json_object_get(iter, "pid"));
			window->focus = json_boolean_value(
				json_object_get(iter, "focus"));
			window->maximized = json_integer_value(
				json_object_get(iter, "maximized"));

			wl_window_pretty(window);
		}
		break;
	default:
	}
}

#ifdef json_array_foreach
#undef json_array_foreach

#define json_array_foreach(array, block)                              \
	json_t *iter;                                                 \
	for (size_t index = 0; index < json_array_size(array) &&      \
			       (iter = json_array_get(array, index)); \
	     index++)                                                 \
	block

#endif

static void json_array_foreach_do(json_t *array,
				  void (*json_decode)(json_t *, void *))
{
	json_array_foreach (array, {
		struct wl_window_desc window;

		json_decode(iter, &window);
	})
}

static json_t *json_load(const char *buf)
{
	json_t *array	 = NULL;
	json_error_t err = {};

	array = json_loads(buf, 0, &err);
	if (!array) {
		fprintf(stderr, "json_loads error: `%s`\n", err.text);
		abort();
	}

	return array;
}

static void __unused list_active_windows(struct dbus_context *ctx)
{
	DBusMessage *msg;
	DBusMessageIter iter;

	msg = dbus_message_new_method_call(DBUS_NAME, DBUS_PATH, DBUS_INTERFACE,
					   "List");
	if (!msg) {
		fprintf(stderr, "Message Null\n");
		exit(1);
	}
	msg = dbus_call_method(ctx, msg);

	dbus_message_iter_init(msg, &iter);

	do {
		int type = dbus_message_iter_get_arg_type(&iter);

		switch (type) {
		case DBUS_TYPE_STRING: {
			const char *buf;
			json_t *array;

			dbus_message_iter_get_basic(&iter, &buf);
			array = json_load(buf);

			json_array_foreach_do(array, json_decode_wl_window);

			break;
		}
		default:
			printf("message type: %c\n", type);
			break;
		}
		dbus_message_iter_next(&iter);

	} while (dbus_message_iter_has_next(&iter));

	dbus_message_unref(msg);
}

static void raise_emacs_window(struct dbus_context *ctx)
{
	DBusMessage *msg;

	msg = dbus_message_new_method_call(DBUS_NAME, DBUS_PATH, DBUS_INTERFACE,
					   "RaiseEmacsWindow");
	if (!msg) {
		fprintf(stderr, "null message: %m\n");
		exit(1);
	}
	dbus_call_method(ctx, msg);
}

int main(void)
{
	struct dbus_context ctx;

	dbus_init(&ctx);

	raise_emacs_window(&ctx);

	dbus_close(&ctx);

	return 0;
}
