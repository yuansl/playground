#include <stdio.h>
#include <stdlib.h>
#include <time.h>

#include <pthread.h>

#define NUM_THREADS 4
#define SLOT_SIZE 240

static inline char *now(void)
{
	static char buf[16 << 10];
	struct timespec t = {};

	clock_gettime(CLOCK_REALTIME, &t);
	strftime(buf, sizeof(buf), "%FT%H:%M:%S%z", localtime(&t.tv_sec));
	return buf;
}

#define print(...)                                                 \
	({                                                         \
		printf("%s: thread %ld: ", now(), pthread_self()); \
		printf(__VA_ARGS__);                               \
	})

struct thread_context {
	pthread_mutex_t *mutex;
	pthread_cond_t *wait_producer;
	pthread_cond_t *wait_consumer;
	int emptyslots;
	int size;
	int head, tail;
	int slots[SLOT_SIZE];
	bool done;
};

#define WITH_LOCKED(mutex, BODY)             \
	{                                    \
		pthread_mutex_lock(mutex);   \
		do {                         \
			BODY;                \
		} while (0);                 \
		pthread_mutex_unlock(mutex); \
	}

void *consumer_routine(void *arg)
{
	struct thread_context *ctx = arg;
	bool stop = false;

	print("consumer starting...\n");

	while (!stop) {
		WITH_LOCKED(ctx->mutex, {
			while (ctx->size - ctx->emptyslots == 0 && !ctx->done) {
				print("consumer going to waiting, head=%d, tail=%d, done=%s\n",
				      ctx->head, ctx->tail,
				      ctx->done ? "true" : "false");

				pthread_cond_wait(ctx->wait_producer,
						  ctx->mutex);
			}
			if (ctx->done && (ctx->size - ctx->emptyslots == 0)) {
				stop = true;
				break;
			}

			bool has_waits = ctx->emptyslots == 0;
			int value = ctx->slots[ctx->head];

			print("consumer running, got slot[%d] = %d\n",
			      ctx->head, value);

			ctx->head = (ctx->head + 1) % ctx->size;
			ctx->emptyslots++;
			if (has_waits)
				pthread_cond_signal(ctx->wait_consumer);
		});
	}

	print("consumer quitting...\n");

	pthread_exit(NULL);
}

int main(void)
{
	pthread_t threads[NUM_THREADS];
	pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;
	pthread_cond_t condvar = PTHREAD_COND_INITIALIZER;
	pthread_cond_t condvar2 = PTHREAD_COND_INITIALIZER;
	struct thread_context ctx = {
		.mutex = &mutex,
		.wait_producer = &condvar,
		.wait_consumer = &condvar2,
		.emptyslots = SLOT_SIZE,
		.size = SLOT_SIZE,
	};
	now();
	struct timespec time;
	clock_gettime(CLOCK_REALTIME, &time);

	srand(time.tv_nsec + time.tv_sec * 1e9);

	for (int i = 0; i < NUM_THREADS; i++) {
		int err;

		err = pthread_create(&threads[i], NULL, consumer_routine, &ctx);
		if (err)
			perror("pthread_create");
	}

	for (int i = 0; i < SLOT_SIZE; i++) {
		int value = rand();

		WITH_LOCKED(ctx.mutex, {
			while (ctx.emptyslots == 0)
				pthread_cond_wait(ctx.wait_consumer, ctx.mutex);

			bool has_waits = ctx.size - ctx.emptyslots == 0;

			print("producer put[%d]: %d\n", ctx.tail, value);

			ctx.slots[ctx.tail++] = value;
			ctx.tail %= ctx.size;
			ctx.emptyslots--;

			if (has_waits)
				pthread_cond_broadcast(ctx.wait_producer);
		});
	}
	WITH_LOCKED(ctx.mutex, {
		bool has_waits = ctx.size - ctx.emptyslots == 0;
		if (has_waits)
			pthread_cond_broadcast(ctx.wait_producer);
		ctx.done = true;
	});
	print("send task done\n");

	for (int i = 0; i < NUM_THREADS; i++) {
		pthread_join(threads[i], NULL);
	}
	return 0;
}
