package taskqueue

import (
	"context"
	"log"
	"sync"
	"time"
)

// Worker represents a single worker that processes tasks from a broker.
// It is responsible for consuming tasks, executing handlers, and managing retries.
type Worker struct {
	id       int                        // The unique identifier for the worker.
	broker   Broker                     // The broker from which tasks are consumed.
	backoff  *BackoffPolicy             // The backoff policy for retrying tasks in case of failure.
	handlers map[string]TaskHandlerFunc // A map of task names to their corresponding handler functions.
	wg       *sync.WaitGroup            // A WaitGroup to synchronize worker shutdown and wait for all tasks to finish.
}

// NewWorker creates a new worker instance with the given ID, broker, backoff policy, and WaitGroup.
// The worker is responsible for consuming tasks from the broker, executing the appropriate handlers,
// and handling retries in case of failures.
func NewWorker(id int, broker Broker, backoff *BackoffPolicy, wg *sync.WaitGroup) *Worker {
	return &Worker{
		id:       id,
		broker:   broker,
		backoff:  backoff,
		handlers: make(map[string]TaskHandlerFunc),
		wg:       wg,
	}
}

// Register registers a task handler function for a specific task name.
// The handler will be invoked when a task with the specified name is consumed from the broker.
func (w *Worker) Register(name string, handler TaskHandlerFunc) {
	w.handlers[name] = handler
}

// Start begins the worker's task-consuming loop. It continuously consumes tasks from the broker and
// invokes the registered task handlers. If a task handler returns an error, the task may be retried
// based on the backoff policy and retry logic. The method ensures graceful shutdown using a context and WaitGroup.
func (w *Worker) Start(ctx context.Context) {
	defer w.wg.Done()

	tasks, err := w.broker.Consume()
	if err != nil {
		log.Fatalf("Failed to consume tasks: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down...")
			return

		case task, ok := <-tasks:
			if !ok {
				log.Println("Task channel closed, worker exiting...")
				return
			}

			handler, ok := w.handlers[task.Name]
			if !ok {
				log.Println("No handler for task", task.Name)
				continue
			}

			err := handler(task.Args)
			if err != nil {
				if task.Retry < task.MaxRetry {
					task.Retry++

					if w.backoff != nil {
						backoffDelay := w.backoff.Calculate(task.Retry)
						log.Printf("Task %s failed, backing off for %v (retry %d/%d)\n", task.Name, backoffDelay, task.Retry, task.MaxRetry)
						time.Sleep(backoffDelay)
					} else {
						log.Printf("Task %s failed, retrying immediately (retry %d/%d)\n", task.Name, task.Retry, task.MaxRetry)
					}

					w.broker.Publish(task)
				} else {
					log.Printf("Worker %d: task %s exceeded retries\n", w.id, task.Name)
				}
			} else {
				log.Printf("Worker %d: task %s succeeded\n", w.id, task.Name)
			}
		}
	}
}
