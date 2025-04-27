package taskqueue

type Broker interface {
	// Publish sends a task to the broker.
	// Returns an error if the task could not be published.
	Publish(task Task) error

	// Consume returns a channel of tasks that the worker can consume.
	// It also returns an error if the broker fails to start consuming tasks.
	Consume() (<-chan Task, error)
}
