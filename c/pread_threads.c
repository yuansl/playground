#include <stdatomic.h>
#include <sys/stat.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <fcntl.h>
#include <time.h>
#include <unistd.h>

#include <pthread.h>

#define BUFSIZE (16 << 10)

struct task {
	int fd;
	off_t off;
	int id;
};

struct task_context {
	pthread_cond_t *consumer_wait;
	pthread_cond_t *producer_wait;
	pthread_mutex_t *mux;
	struct task **task_slots;
	int nr_slots;
	int producer;
	int consumer;
	atomic_size_t nbytes;
	atomic_size_t size;
};

void *pread_rotine(void *args)
{
	struct task_context *context = args;
	struct task *task;
	char buf[BUFSIZE];
	ssize_t n;

	while (atomic_load(&context->nbytes) < atomic_load(&context->size)) {
		pthread_mutex_lock(context->mux);

		while (context->consumer >= context->producer) {
			printf("thread %ld waiting on condvar\n",
			       pthread_self());
			pthread_cond_wait(context->producer_wait, context->mux);
		}

		task = context->task_slots[context->consumer];

		n = pread(task->fd, buf, sizeof(buf), task->off);
		if (n < 0) {
			fprintf(stderr, "pread: %m\n");
			pthread_exit(NULL);
		}
		printf("#%d thread %ld running: fd=%d,off=%ld,read %zd bytes\n",
		       task->id, pthread_self(), task->fd, task->off, n);

		context->consumer = (context->consumer + 1) % context->nr_slots;
		pthread_mutex_unlock(context->mux);
		atomic_fetch_add(&context->nbytes, n);

		pthread_cond_broadcast(context->consumer_wait);
	}

	pthread_exit(NULL);
}

#define NUM_CPUS 4

static void pread_in_parallel(int fd)
{
	struct stat stat = {};
	int nr_task;
	struct task_context shared;
	ssize_t nbytes;
	clock_t start, stop;
	int ncpus = sysconf(_SC_NPROCESSORS_ONLN);
	pthread_t threads[ncpus];
	pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;
	pthread_cond_t wait_consumer = PTHREAD_COND_INITIALIZER;
	pthread_cond_t wait_producer = PTHREAD_COND_INITIALIZER;

	if (ncpus < NUM_CPUS)
		ncpus = NUM_CPUS;

	printf("ncpus=%d\n", ncpus);

	fstat(fd, &stat);
	nr_task = stat.st_size / BUFSIZE;
	if (stat.st_size % BUFSIZE) {
		nr_task += 1;
	}

	printf("number of tasks: %d, sizeof(struct task)=%zd, online numcpu: %d\n",
	       nr_task, sizeof(struct task), ncpus);

	shared = (struct task_context){
		.mux = &mutex,
		.consumer_wait = &wait_consumer,
		.producer_wait = &wait_producer,
		.consumer = 0,
		.producer = 0,
		.size = stat.st_size,
		.nbytes = 0,
		.nr_slots = ncpus,
		.task_slots = calloc(ncpus, sizeof(struct task *)),
	};
	if (!shared.task_slots)
		perror("calloc");

	start = clock();
	for (int i = 0; i < ncpus; i++) {
		int err = pthread_create(&threads[i], NULL, pread_rotine,
					 &shared);
		if (err) {
			fprintf(stderr, "pthread_create: %s\n", strerror(err));
			exit(3);
		}
	}
	for (off_t off = 0, i = 0; off < stat.st_size; off += BUFSIZE, i++) {
		struct task *task = calloc(1, sizeof(*task));
		task = &(struct task){ .fd = fd, .off = off, .id = i };

		printf("producer: putting task ...\n");
		pthread_mutex_lock(shared.mux);

		printf("producer: put a task done\n");
		shared.task_slots[shared.producer] = task;

		int next = (shared.producer + 1) % shared.nr_slots;

		while (next == shared.consumer)
			pthread_cond_wait(shared.consumer_wait, shared.mux);
		shared.producer = next;
		pthread_mutex_unlock(shared.mux);

		pthread_cond_broadcast(shared.producer_wait);
	}

	nbytes = 0;
	for (int i = 0; i < ncpus; i++) {
		pthread_join(threads[i], NULL);
	}
	stop = clock();

	printf("read %zd bytes in total, elapsed time: %.2f sec\n", nbytes,
	       (double)(stop - start) / CLOCKS_PER_SEC);

	for (int i = 0; i < ncpus; i++) {
		free(shared.task_slots[i]);
	}
	free(shared.task_slots);
}

int main(int argc, char *argv[])
{
	int fd;

	if (argc < 2) {
		fprintf(stderr, "Usage: %s <filename>\n", argv[0]);
		exit(1);
	}

	fd = open(argv[1], O_RDONLY);
	if (fd < 0) {
		fprintf(stderr, "open: %m\n");
		exit(2);
	}

	pread_in_parallel(fd);

	return 0;
}
