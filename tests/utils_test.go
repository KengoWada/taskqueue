package taskqueue_test

import (
	"context"
	"errors"
	"sync"

	"github.com/KengoWada/taskqueue"
)

var errSimulated = errors.New("simulated error")

type MockWorker struct {
	id      int
	Handler taskqueue.HandlerRegistry
	Started bool
	wg      *sync.WaitGroup
}

func (w *MockWorker) Register(name string, handler taskqueue.TaskHandlerFunc) {
	w.Handler[name] = handler
}

func (w *MockWorker) Start(ctx context.Context) {
	w.Started = true
	defer func() { w.Started = false }()
	defer w.wg.Done()

	<-ctx.Done()
}

func MockWorkerFactory(cfg taskqueue.WorkerConfig) taskqueue.Worker {
	return &MockWorker{
		id:      cfg.ID,
		wg:      cfg.WG,
		Handler: make(taskqueue.HandlerRegistry),
		Started: false,
	}
}

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
