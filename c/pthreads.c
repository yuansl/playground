#include <bits/time.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <time.h>

#define SLOT_SIZE 18

typedef intmax_t timestamp_t;

static inline char *now(void)
{
	static char buf[16 << 10];
	struct timespec t = {};

	clock_gettime(CLOCK_REALTIME, &t);
	strftime(buf, sizeof(buf), "%FT%H:%M:%S%z", localtime(&t.tv_sec));
	return buf;
}

struct thread_context {
	pthread_mutex_t *mutex;
	pthread_cond_t *condvar;
	int head, tail;
	int slots[SLOT_SIZE];
	bool done;
};

#define print(...)                     \
	({                             \
		printf("%s: ", now()); \
		printf(__VA_ARGS__);   \
	})

void *pthread_routine(void *arg)
{
	struct thread_context *ctx = arg;

	print("thread %ld running...\n", pthread_self());

	while (true) {
		pthread_mutex_lock(ctx->mutex);
		while (ctx->head == ctx->tail && !ctx->done) {
			print("thread %ld waiting, head=%d, tail=%d, done=%s\n",
			      pthread_self(), ctx->head, ctx->tail,
			      ctx->done ? "true" : "false");

			pthread_cond_wait(ctx->condvar, ctx->mutex);
		}
		if (ctx->done)
			break;

		print("thread %ld running, got slot = %d, head=%d, tail=%d\n",
		      pthread_self(), ctx->slots[ctx->head], ctx->head,
		      ctx->tail);

		ctx->head = (ctx->head + 1) % SLOT_SIZE;

		pthread_mutex_unlock(ctx->mutex);
	}

	print("thread %ld quitting...\n", pthread_self());

	pthread_exit(NULL);
}

#define NUM_THREADS 2

pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;
pthread_cond_t condvar = PTHREAD_COND_INITIALIZER;

int main(void)
{
	pthread_t threads[NUM_THREADS];
	struct thread_context ctx = {
		.mutex = &mutex,
		.condvar = &condvar,
		.done = false,
	};

	srand(time(NULL));

	for (int i = 0; i < NUM_THREADS; i++) {
		int err;

		err = pthread_create(&threads[i], NULL, pthread_routine, &ctx);
		if (err)
			perror("pthread_create");
	}

	for (int i = 0; i < SLOT_SIZE; i++) {
		pthread_mutex_lock(ctx.mutex);

		ctx.slots[ctx.tail] = rand() % 1000;

		int next = (ctx.tail + 1) % SLOT_SIZE;
		while (next == ctx.head) {
			print("producer waiting for condition\n");
			pthread_cond_wait(ctx.condvar, ctx.mutex);
		}

		ctx.tail = next;

		print("put slot head=%d, tail=%d\n", ctx.head, ctx.tail);
		pthread_mutex_unlock(ctx.mutex);

		pthread_cond_broadcast(ctx.condvar);
	}
	pthread_mutex_lock(ctx.mutex);
	ctx.done = true;
	pthread_mutex_unlock(ctx.mutex);

	print("task done(nanoseconds)\n");

	for (int i = 0; i < NUM_THREADS; i++) {
		pthread_join(threads[i], NULL);
	}

	return 0;
}
