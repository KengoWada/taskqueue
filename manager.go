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
	BackoffPolicy *BackoffPolicy
}

// Manager is responsible for managing the lifecycle of workers,
// coordinating task consumption and processing, and gracefully stopping workers.
type Manager struct {
	broker  Broker
	workers []*Worker

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
func WithBackoffPolicy(bp *BackoffPolicy) ManagerOption {
	return func(cfg *ManagerConfig) {
		cfg.BackoffPolicy = bp
	}
}

// NewManager creates and returns a new Manager instance. The Manager is responsible for
// managing workers that process tasks from the provided Broker. It allows for a custom
// number of workers and optional configuration such as a backoff policy for task retries.
//
// The function accepts a Broker, the number of workers to create, and an optional list of
// ManagerOption functions for additional configuration. If no backoff policy is provided,
// a default backoff policy will be applied.
//
// Parameters:
//   - broker: The Broker to be used for publishing and consuming tasks.
//   - numWorkers: The number of workers to be created (defaults to 1 if less than or equal to 0).
//   - opts: A variadic list of ManagerOption functions for additional configuration.
//
// Returns:
//   - A pointer to the created Manager instance.
//
// Example usage:
//
//	broker := NewRabbitMQBroker(...)
//	manager := NewManager(broker, 3, WithBackoffPolicy(customBackoffPolicy))
func NewManager(broker Broker, numWorkers int, opts ...ManagerOption) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	if numWorkers <= 0 {
		numWorkers = DefaultNumOfWorkers
	}

	cfg := &ManagerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Set default backoff if cfg.BackoffPolicy == nil
	if cfg.BackoffPolicy == nil {
		cfg.BackoffPolicy = &BackoffPolicy{
			BaseDelay:     1 * time.Second,
			MaxDelay:      30 * time.Second,
			UseJitter:     true,
			JitterRangeMs: 300,
		}
	}

	m := &Manager{broker: broker, ctx: ctx, cancel: cancel}
	for i := 1; i <= numWorkers; i++ {
		m.wg.Add(1)
		worker := NewWorker(i, broker, cfg.BackoffPolicy, &m.wg)
		m.workers = append(m.workers, worker)
	}

	return m
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
