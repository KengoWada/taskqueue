package taskqueue_test

import (
	"testing"
	"time"

	"github.com/KengoWada/taskqueue"
	"github.com/stretchr/testify/assert"
)

func TestManager(t *testing.T) {
	t.Run("should initialize manager with workers and default backoff", func(t *testing.T) {
		mockBroker := NewMockBroker(1)
		tests := []struct {
			numOfWorkers int
			expected     int
			setBackoff   bool
		}{
			{numOfWorkers: -1, expected: 1, setBackoff: true},
			{numOfWorkers: 0, expected: 1, setBackoff: false},
			{numOfWorkers: 5, expected: 5, setBackoff: true},
			{numOfWorkers: 10, expected: 10, setBackoff: false},
		}

		backoffPolicy := &taskqueue.BackoffPolicy{
			BaseDelay:     5 * time.Millisecond, // small delay for fast test
			MaxDelay:      20 * time.Millisecond,
			UseJitter:     false,
			JitterRangeMs: 0,
		}

		for _, tt := range tests {
			var manager *taskqueue.Manager
			var expectedBackoff *taskqueue.BackoffPolicy
			if tt.setBackoff {
				manager = taskqueue.NewManager(mockBroker, MockWorkerFactory, tt.numOfWorkers, taskqueue.WithBackoffPolicy(backoffPolicy))
				expectedBackoff = backoffPolicy
			} else {
				manager = taskqueue.NewManager(mockBroker, MockWorkerFactory, tt.numOfWorkers)
				expectedBackoff = &taskqueue.DefaultBackoffPolicy
			}

			assert.Equal(t, tt.expected, len(manager.Workers()))
			assert.Equal(t, expectedBackoff, manager.BackoffPolicy().(*taskqueue.BackoffPolicy))
		}
	})

	t.Run("should register handler with all workers", func(t *testing.T) {
		mockBroker := NewMockBroker(1)
		numOfWorkers := 5
		manager := taskqueue.NewManager(mockBroker, MockWorkerFactory, numOfWorkers)

		taskName := "task_name"
		manager.RegisterTask(taskName, func(ta taskqueue.TaskArgs) error { return nil })

		workers := manager.Workers()
		assert.Equal(t, numOfWorkers, len(workers))

		for _, worker := range workers {
			w := worker.(*MockWorker)
			task, exists := w.Handler[taskName]
			assert.True(t, exists)
			assert.Nil(t, task(taskqueue.TaskArgs{}))
		}
	})

	t.Run("should publish task to the broker", func(t *testing.T) {
		mockBroker := NewMockBroker(1)
		numOfWorkers := 5
		manager := taskqueue.NewManager(mockBroker, MockWorkerFactory, numOfWorkers)

		// Invalid task name as empty string
		err := manager.PublishTask("", taskqueue.TaskArgs{}, 3)
		assert.Equal(t, taskqueue.ErrEmptyTaskName, err)

		// Valid task name and arguments
		err = manager.PublishTask("task_name", taskqueue.TaskArgs{}, 2)
		assert.Nil(t, err)

		// Broker fails to publish task
		mockBroker.badPublish = true
		err = manager.PublishTask("task_name", taskqueue.TaskArgs{}, 2)
		assert.NotNil(t, err)
		assert.Equal(t, errSimulated, err)
	})

	t.Run("should start and stop all workers", func(t *testing.T) {
		mockBroker := NewMockBroker(1)
		numOfWorkers := 5
		manager := taskqueue.NewManager(mockBroker, MockWorkerFactory, numOfWorkers)

		manager.Start()

		time.Sleep(100 * time.Millisecond)

		workers := manager.Workers()
		assert.Equal(t, numOfWorkers, len(workers))
		for _, worker := range workers {
			w := worker.(*MockWorker)
			assert.True(t, w.Started)
		}

		manager.Stop()

		time.Sleep(100 * time.Millisecond)

		for _, worker := range workers {
			w := worker.(*MockWorker)
			assert.False(t, w.Started)
		}
	})
}
