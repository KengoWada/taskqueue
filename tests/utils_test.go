package taskqueue_test

import (
	"errors"
	"sync"

	"github.com/KengoWada/taskqueue"
)

var errSimulated = errors.New("simulated error")

type MockBroker struct {
	mu             sync.Mutex
	tasks          chan taskqueue.Task
	publishedTasks []taskqueue.Task
	badConsume     bool
	badPublish     bool
	closeChannel   bool
}

func NewMockBroker(buffer int) *MockBroker {
	return &MockBroker{
		tasks: make(chan taskqueue.Task, buffer),
	}
}

func (m *MockBroker) Publish(task taskqueue.Task) error {
	if m.badPublish {
		return errSimulated
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedTasks = append(m.publishedTasks, task)
	m.tasks <- task
	return nil
}

func (m *MockBroker) Consume() (<-chan taskqueue.Task, error) {
	if m.badConsume {
		return nil, errSimulated
	}

	if m.closeChannel {
		close(m.tasks)
	}
	return m.tasks, nil
}
