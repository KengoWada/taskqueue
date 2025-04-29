package taskqueue

import (
	"context"
	"log"
	"sync"
	"time"
)

// Worker defines the contract for a task-processing worker.
//
// Implementations of Worker are responsible for registering task handlers
// and starting the task execution loop, typically consuming tasks from a Broker.
//
// Methods:
//   - Register: Associates a task name with a handler function.
//   - Start: Begins processing tasks using the provided context for cancellation.
//
// This interface allows different worker implementations to be plugged into the Manager,
// enabling customizable behavior such as logging, metrics, or concurrency models.
//
// Example usage:
//
//	type MyWorker struct { ... }
//	func (w *MyWorker) Register(name string, handler TaskHandlerFunc) { ... }
//	func (w *MyWorker) Start(ctx context.Context) { ... }
type Worker interface {
	Register(name string, handler TaskHandlerFunc)
	Start(ctx context.Context)
}

// WorkerFactory defines a function that creates a new Worker instance using the provided configuration.
//
// This allows users to customize how Workers are constructed, enabling injection of custom
// dependencies (e.g., loggers, metrics, settings) without modifying the Manager.
//
// The WorkerConfig contains common fields such as worker ID, broker, backoff policy, and a wait group.
//
// Example:
//
//	func MyWorkerFactory(cfg WorkerConfig) Worker {
//	    return &MyCustomWorker{
//	        id:      cfg.ID,
//	        broker:  cfg.Broker,
//	        backoff: cfg.Backoff,
//	        wg:      cfg.WG,
//	        logger:  myLogger, // custom dependency
//	    }
//	}
type WorkerFactory func(cfg WorkerConfig) Worker

// WorkerConfig provides the configuration necessary to initialize a Worker.
//
// It is passed to the WorkerFactory function by the Manager to ensure that all workers
// have the required shared dependencies and context-specific data.
//
// Fields:
//   - ID: A unique identifier for the worker, typically assigned by the Manager.
//   - Broker: The Broker instance used to retrieve and dispatch tasks.
//   - Backoff: The policy used for retrying tasks on failure.
//   - WG: A shared WaitGroup used by the Manager to coordinate worker shutdown.
//
// This struct is designed to be extended if additional shared dependencies need to
// be passed to workers in the future.
type WorkerConfig struct {
	ID      int
	Broker  Broker
	Backoff *BackoffPolicy
	WG      *sync.WaitGroup
}

// DefaultWorker is the standard implementation of the Worker interface.
//
// It is responsible for consuming tasks from the provided Broker, executing them using
// registered handler functions, and retrying failed tasks based on a backoff policy.
//
// Fields:
//   - id: A unique identifier for this worker instance.
//   - broker: The Broker used to fetch tasks for execution.
//   - backoff: The BackoffPolicy used to delay retries on task failure.
//   - handlers: A map of task names to their associated TaskHandlerFunc implementations.
//   - wg: A shared WaitGroup used by the Manager to coordinate worker shutdown.
//
// DefaultWorker is intended to be created using a WorkerFactory and managed by a Manager.
// It supports context-based cancellation for graceful shutdown.
type DefaultWorker struct {
	id       int
	broker   Broker
	backoff  *BackoffPolicy
	handlers map[string]TaskHandlerFunc
	wg       *sync.WaitGroup
}

// DefaultWorkerFactory creates and returns a new instance of DefaultWorker using the provided WorkerConfig.
//
// The worker is initialized with:
//   - A unique ID
//   - A broker for task consumption
//   - A backoff policy for retrying failed tasks
//   - A shared WaitGroup for graceful shutdown coordination
//   - An empty handler map ready for task registration
//
// This function returns a Worker interface, allowing the DefaultWorker to be used polymorphically.
//
// Typically used within a WorkerFactory passed to the Manager.
func DefaultWorkerFactory(cfg WorkerConfig) Worker {
	return &DefaultWorker{
		id:       cfg.ID,
		broker:   cfg.Broker,
		backoff:  cfg.Backoff,
		handlers: make(map[string]TaskHandlerFunc),
		wg:       cfg.WG,
	}
}

// Register registers a task handler for a specific task name.
//
// The handler function will be associated with the provided task name and
// invoked when a task with that name is received by the worker.
//
// Parameters:
//   - name: The unique name of the task being registered.
//   - handler: The TaskHandlerFunc that will handle the task when it is dispatched.
//
// This method allows users to dynamically register multiple tasks for a worker,
// enabling the worker to handle a variety of task types.
//
// Example:
//
//	worker.Register("send_email", sendEmailHandler)
func (w *DefaultWorker) Register(name string, handler TaskHandlerFunc) {
	w.handlers[name] = handler
}

// Start begins processing tasks for the worker and continuously listens for new tasks
// from the broker until the provided context is cancelled or the task channel is closed.
//
// It runs a task consumption loop that will:
//   - Consume tasks from the broker
//   - Attempt to handle each task using the corresponding registered handler
//   - Retry failed tasks according to the worker's backoff policy (if provided)
//   - Exit when the context is cancelled or the task channel is closed
//
// Parameters:
//   - ctx: The context used to control the lifetime of the worker. When the context
//     is cancelled, the worker will shut down gracefully.
//
// The worker will log information about task successes, failures, retries, and backoff delays.
//
// Example:
//
//	worker.Start(ctx)
func (w *DefaultWorker) Start(ctx context.Context) {
	defer w.wg.Done()

	tasks, err := w.broker.Consume()
	if err != nil {
		log.Printf("Failed to consume tasks: %v\n", err)
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
