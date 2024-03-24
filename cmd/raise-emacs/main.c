#include "dbus/dbus-protocol.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <jansson.h>
#include <dbus/dbus.h>

#define DB_DESTINATION "org.gnome.Shell"
#define DB_PATH	       "/org/gnome/Shell/Extensions/WindowsExt"
#define DB_INTERFACE   "org.gnome.Shell.Extensions.WindowsExt"

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

static DBusMessage *dbus_call_method(struct dbus_context *ctx, DBusMessage *msg)
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

#ifdef json_array_foreach
#undef json_array_foreach
#define json_array_foreach(array, elem)                               \
	for (size_t index = 0; index < json_array_size(array) &&      \
			       (elem = json_array_get(array, index)); \
	     index++)

#endif

static void json_array_foreach_do(json_t *array)
{
	json_t *iter;

	json_array_foreach (array, iter) {
		json_t *class = json_object_get(iter, "class");
		switch ((int)class->type) {
		case JSON_STRING:
			if (!strcmp(json_string_value(class), "emacs")) {
				json_t *title = json_object_get(iter, "title");
				json_t *pid   = json_object_get(iter, "pid");
				json_t *id    = json_object_get(iter, "id");
				json_t *focus = json_object_get(iter, "focus");
				json_t *maximized =
					json_object_get(iter, "maximized");
				printf("title=%s,pid=%lld,id=%lld,focus=%s,maximized=%lld\n",
				       json_string_value(title),
				       json_integer_value(pid),
				       json_integer_value(id),
				       json_boolean_value(focus) ? "true" :
								   "false",
				       json_integer_value(maximized));
			}
			break;
		default:
		}
	}
}

[[maybe_unused]] static void list_active_windows(struct dbus_context *ctx)
{
	DBusMessage *msg;
	DBusMessageIter iter;

	msg = dbus_message_new_method_call(DB_DESTINATION, // target for
							   // the method
							   // call
					   DB_PATH,	   // object to call on
					   DB_INTERFACE,   // interface to
							 // call on
					   "List"); // method name
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
			dbus_message_iter_get_basic(&iter, &buf);
			json_error_t err;
			json_t *array = json_loads(buf, 0, &err);

			if (!array) {
				fprintf(stderr, "json_loads error: `%s`\n",
					err.text);
				abort();
			}
			json_array_foreach_do(array);

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

	msg = dbus_message_new_method_call(DB_DESTINATION, // target for
							   // the method
							   // call
					   DB_PATH,	   // object to call on
					   DB_INTERFACE,   // interface to
							 // call on
					   "RaiseEmacsWindow"); // method name
	if (!msg) {
		fprintf(stderr, "Message Null\n");
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
