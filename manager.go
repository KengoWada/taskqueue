package taskqueue

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

const DefaultNumOfWorkers int = 1

// ManagerOption is a function type that modifies the configuration of a ManagerConfig.
// It allows for functional configuration of the Manager, enabling users to customize
// various settings, such as the backoff policy, when creating a Manager.
type ManagerOption func(*ManagerConfig)

// ManagerConfig holds configuration options for the Manager.
type ManagerConfig struct {
	BackoffPolicy Backoff
}

// Manager is responsible for managing the lifecycle of workers,
// coordinating task consumption and processing, and gracefully stopping workers.
type Manager struct {
	broker  Broker
	workers []Worker

	ctx    context.Context    // The context used for managing the lifecycle of the Manager.
	cancel context.CancelFunc // The cancel function associated with the context, used to stop the Manager.
	wg     sync.WaitGroup     // A WaitGroup to wait for all workers to finish their tasks.
}

// WithBackoffPolicy is a functional option that sets the BackoffPolicy for a Manager.
// It allows the user to specify a custom backoff policy for retrying failed tasks in the Manager.
//
// Example usage:
//
//	manager := NewManager(bp, numOfWorkers, WithBackoffPolicy(customBackoffPolicy))
func WithBackoffPolicy(bp Backoff) ManagerOption {
	return func(cfg *ManagerConfig) {
		cfg.BackoffPolicy = bp
	}
}

// NewManager creates a new Manager instance responsible for coordinating and supervising workers.
//
// Parameters:
//   - broker: The Broker used by all workers to fetch and process tasks.
//   - wf: A WorkerFactory function used to create each Worker with a provided WorkerConfig.
//   - numWorkers: The number of workers to spawn. If set to 0 or less, DefaultNumOfWorkers is used.
//   - opts: Optional functional configuration parameters for customizing Manager behavior (e.g. backoff policy).
//
// The Manager sets up a cancellable context for controlling the lifecycle of its workers,
// and assigns a shared WaitGroup to coordinate shutdowns. If no BackoffPolicy is provided in
// the options, a DefaultBackoffPolicy is used.
//
// Each worker is created using the WorkerFactory and given a unique ID, shared broker,
// backoff policy, and reference to the Manager's WaitGroup.
//
// Returns:
//
//	A pointer to a fully initialized Manager ready to register task handlers and start processing.
//
// Example:
//
//	mgr := NewManager(broker, MyWorkerFactory(), 5, WithBackoff(customBackoff))
//	mgr.RegisterTask("send_email", emailHandler)
//	mgr.Start()
func NewManager(broker Broker, wf WorkerFactory, numWorkers int, opts ...ManagerOption) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	if numWorkers <= 0 {
		numWorkers = DefaultNumOfWorkers
	}

	managerConfig := &ManagerConfig{}
	for _, opt := range opts {
		opt(managerConfig)
	}

	if managerConfig.BackoffPolicy == nil {
		managerConfig.BackoffPolicy = &DefaultBackoffPolicy
	}

	manager := &Manager{broker: broker, ctx: ctx, cancel: cancel}
	for i := range numWorkers {
		manager.wg.Add(1)
		workerConfig := WorkerConfig{
			ID:      i + 1,
			Broker:  manager.broker,
			Backoff: managerConfig.BackoffPolicy,
			WG:      &manager.wg,
		}
		manager.workers = append(manager.workers, wf(workerConfig))
	}

	return manager
}

// RegisterTask registers a task handler for the specified task name across all workers managed by the Manager.
// This allows each worker to handle tasks of the specified name by executing the provided handler function.
//
// The handler function must have the signature: func(TaskArgs) error. If the task name is already registered,
// the handler will overwrite the existing one.
//
// Parameters:
//   - taskName: The name of the task to register. This name is used to identify tasks in the queue.
//   - handler: The function that handles the task when it is consumed by the worker.
//
// Example usage:
//
//	manager.RegisterTask("send_email", func(args TaskArgs) error {
//	    // Handle the task
//	    return nil
//	})
func (m *Manager) RegisterTask(taskName string, handler TaskHandlerFunc) {
	for _, w := range m.workers {
		w.Register(taskName, handler)
	}
}

// PublishTask publishes a task to the broker with the specified task name, arguments, and maximum retries.
//
// This method creates a new task with the provided parameters and publishes it to the broker for consumption by workers.
//
// Parameters:
//   - taskName: The name of the task to publish. This is used to identify the task in the queue.
//   - args: The arguments that will be passed to the task handler. These are passed as a map of key-value pairs.
//   - maxRetry: The maximum number of retries allowed for the task in case of failure. If the task fails more than
//     the specified number of times, it will not be retried again.
//
// Returns:
//   - An error if the task name is empty or there is an issue publishing the task to the broker. If no errors occur,
//     the method will return nil.
//
// Example usage:
//
//	err := manager.PublishTask("send_email", TaskArgs{"email": "user@example.com"}, 3)
//	if err != nil {
//	    log.Printf("Error publishing task: %v", err)
//	}
func (m *Manager) PublishTask(taskName string, args TaskArgs, maxRetry int) error {
	if taskName == "" {
		return errors.New("task name must not be empty")
	}

	task := Task{
		Name:      taskName,
		Args:      args,
		MaxRetry:  maxRetry,
		Timestamp: time.Now().UTC(),
	}

	return m.broker.Publish(task)
}

// Start starts all the workers managed by the Manager in separate goroutines.
//
// This method launches each worker's `Start` method concurrently, passing the Manager's context to each worker.
// The workers will begin consuming tasks from the broker and processing them based on the registered task handlers.
//
// Example usage:
//
//	manager.Start()  // Starts all workers concurrently
func (m *Manager) Start() {
	for _, w := range m.workers {
		go w.Start(m.ctx)
	}
}

// Stop stops the Manager and all its workers gracefully.
//
// This method cancels the context associated with the Manager, signaling all workers to stop their work.
// It then waits for all workers to complete their shutdown process using the WaitGroup.
//
// It is important to call `Stop` to ensure that all workers have finished their tasks and the system shuts down cleanly.
//
// Example usage:
//
//	manager.Stop()  // Stops all workers and waits for them to finish
func (m *Manager) Stop() {
	log.Println("Manager stopping...")
	m.cancel()
	m.wg.Wait()
	log.Println("All workers have shut down.")
}
