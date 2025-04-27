package taskqueue

import "time"

// TaskArgs represents the arguments passed to a task handler.
// It is a map where the key is a string (usually a task argument name)
// and the value is of type `any` which allows flexibility in the type of data.
type TaskArgs map[string]any

// TaskHandlerFunc defines the signature of a function that handles tasks.
// It takes a TaskArgs object as input and returns an error if something
// goes wrong during task processing.
type TaskHandlerFunc func(TaskArgs) error

// Task represents an individual task in the task queue.
type Task struct {
	Name      string    `json:"name"`      // The name of the task, used to identify which handler to invoke.
	Args      TaskArgs  `json:"args"`      // The arguments for the task, provided as a map of key-value pairs.
	Retry     int       `json:"retry"`     // The current retry count for this task.
	MaxRetry  int       `json:"max_retry"` // The maximum number of retries before the task is considered failed.
	Timestamp time.Time `json:"timestamp"` // The timestamp when the task was created or scheduled.
}
