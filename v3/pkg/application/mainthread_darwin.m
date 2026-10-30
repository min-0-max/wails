//go:build darwin && !ios && !server

#import <Foundation/Foundation.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include "mainthread_darwin.h"

// A unit of work that one side hands to the other. The side that hands it over
// keeps it on its stack and waits until done is set.
typedef struct wailsWork {
	void (^block)(void);
	bool done;
	struct wailsWork *next;
} wailsWork;

typedef struct {
	wailsWork *head;
	wailsWork *tail;
} wailsQueue;

// One lock guards both queues and the done flags. The main thread waits on
// wailsMainWake and the worker waits on wailsWorkerWake.
static pthread_mutex_t wailsLock = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t wailsMainWake = PTHREAD_COND_INITIALIZER;
static pthread_cond_t wailsWorkerWake = PTHREAD_COND_INITIALIZER;
// Work that the main thread hands to the worker.
static wailsQueue wailsWorkerQueue;
// Calls that the worker forwards to the main thread.
static wailsQueue wailsMainQueue;
// How many waits of the main thread for the worker are open.
static int wailsMainWaits;

static __thread bool wailsOnWorker;

static void wailsQueuePush(wailsQueue *queue, wailsWork *work) {
	work->next = NULL;
	if (queue->tail == NULL) {
		queue->head = work;
	} else {
		queue->tail->next = work;
	}
	queue->tail = work;
}

static wailsWork *wailsQueuePop(wailsQueue *queue) {
	wailsWork *work = queue->head;
	if (work != NULL) {
		queue->head = work->next;
		if (queue->head == NULL) {
			queue->tail = NULL;
		}
	}
	return work;
}

// wailsServe runs the work of queue until the work until is done, or forever
// when until is NULL. The caller holds wailsLock, and wailsServe returns with
// it held. wake is the condition of the calling side and done is the condition
// of the side that waits for the work that wailsServe runs.
static void wailsServe(wailsQueue *queue, pthread_cond_t *wake, pthread_cond_t *done, wailsWork *until, bool pooled) {
	while (until == NULL || !until->done) {
		wailsWork *work = wailsQueuePop(queue);
		if (work == NULL) {
			pthread_cond_wait(wake, &wailsLock);
			continue;
		}
		pthread_mutex_unlock(&wailsLock);
		if (pooled) {
			@autoreleasepool {
				work->block();
			}
		} else {
			// A forwarded call runs in the autorelease pool of the main thread,
			// so what it returns stays valid until the step that waits for it
			// ends.
			work->block();
		}
		pthread_mutex_lock(&wailsLock);
		work->done = true;
		pthread_cond_broadcast(done);
	}
}

bool wailsIsWorkerThread(void) {
	return wailsOnWorker;
}

void wailsWorkerRun(void) {
	if (pthread_main_np()) {
		fprintf(stderr, "wails: the main-thread worker was started on the main thread\n");
		abort();
	}
	wailsOnWorker = true;
	pthread_setname_np("wails main-thread worker");
	pthread_mutex_lock(&wailsLock);
	wailsServe(&wailsWorkerQueue, &wailsWorkerWake, &wailsMainWake, NULL, true);
}

void wailsRunOnWorker(void (^task)(void)) {
	if (!pthread_main_np()) {
		task();
		return;
	}
	wailsWork work = { .block = task };
	pthread_mutex_lock(&wailsLock);
	wailsQueuePush(&wailsWorkerQueue, &work);
	pthread_cond_broadcast(&wailsWorkerWake);
	wailsMainWaits++;
	wailsServe(&wailsMainQueue, &wailsMainWake, &wailsWorkerWake, &work, false);
	wailsMainWaits--;
	pthread_mutex_unlock(&wailsLock);
}

void wailsRunOnMain(void (^call)(void)) {
	if (!wailsOnWorker) {
		call();
		return;
	}
	wailsWork work = { .block = call };
	pthread_mutex_lock(&wailsLock);
	if (wailsMainWaits == 0) {
		// The worker runs Go code only for work that the main thread waits
		// for, so a forwarded call always has a main thread that runs it.
		pthread_mutex_unlock(&wailsLock);
		fprintf(stderr, "wails: the worker forwarded a call to the main thread while the main thread does not wait for the worker\n");
		abort();
	}
	wailsQueuePush(&wailsMainQueue, &work);
	pthread_cond_broadcast(&wailsMainWake);
	wailsServe(&wailsWorkerQueue, &wailsWorkerWake, &wailsMainWake, &work, true);
	pthread_mutex_unlock(&wailsLock);
}

void wailsRunOnMainSync(void (^call)(void)) {
	if (pthread_main_np()) {
		call();
	} else if (wailsOnWorker) {
		wailsRunOnMain(call);
	} else {
		dispatch_sync(dispatch_get_main_queue(), call);
	}
}
