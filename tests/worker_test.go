package taskqueue_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/KengoWada/taskqueue"
	"github.com/stretchr/testify/assert"
)

func TestWorkerProcesses(t *testing.T) {
	taskName := "test_task"
	taskHandlerCalled := false
	taskHandler := func(args taskqueue.TaskArgs) error {
		taskHandlerCalled = true
		return nil
	}

	failTaskName := "fail_task"
	failCount := 0
	failTaskHandler := func(args taskqueue.TaskArgs) error {
		failCount += 1
		fmt.Println("here")
		return errors.New("simulated error")
	}

	maxRetries := 3

	t.Run("should run task successfully", func(t *testing.T) {
		mockBroker := NewMockBroker(1)
		wg := &sync.WaitGroup{}
		cfg := taskqueue.WorkerConfig{ID: 1, Broker: mockBroker, Backoff: nil, WG: wg}
		worker := taskqueue.DefaultWorkerFactory(cfg)
		worker.Register(taskName, taskHandler)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg.Add(1)
		go worker.Start(ctx)

		task := taskqueue.Task{
			Name: taskName,
			Args: taskqueue.TaskArgs{},
		}
		taskHandlerCalled = false
		err := mockBroker.Publish(task)
		assert.Nil(t, err)

		time.Sleep(100 * time.Millisecond)
		cancel()
		wg.Wait()

		assert.True(t, taskHandlerCalled)
	})

	t.Run("should handle failing task with backoff", func(t *testing.T) {
		backoff := &taskqueue.BackoffPolicy{
			BaseDelay:     5 * time.Millisecond, // small delay for fast test
			MaxDelay:      20 * time.Millisecond,
			UseJitter:     false,
			JitterRangeMs: 0,
		}

		mockBroker := NewMockBroker(5)
		wg := &sync.WaitGroup{}
		cfg := taskqueue.WorkerConfig{ID: 1, Broker: mockBroker, Backoff: backoff, WG: wg}
		worker := taskqueue.DefaultWorkerFactory(cfg)
		worker.Register(failTaskName, failTaskHandler)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg.Add(1)
		go worker.Start(ctx)

		task := taskqueue.Task{
			Name:     failTaskName,
			Args:     taskqueue.TaskArgs{},
			MaxRetry: maxRetries,
		}
		failCount = 0
		err := mockBroker.Publish(task)
		assert.Nil(t, err)

		time.Sleep(100 * time.Millisecond)
		cancel()
		wg.Wait()

		assert.Equal(t, maxRetries+1, failCount)
	})

	t.Run("should handle failing task with no backoff", func(t *testing.T) {
		mockBroker := NewMockBroker(5)
		wg := &sync.WaitGroup{}
		cfg := taskqueue.WorkerConfig{ID: 1, Broker: mockBroker, Backoff: nil, WG: wg}
		worker := taskqueue.DefaultWorkerFactory(cfg)
		worker.Register(failTaskName, failTaskHandler)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg.Add(1)
		go worker.Start(ctx)

		task := taskqueue.Task{
			Name:     failTaskName,
			Args:     taskqueue.TaskArgs{},
			MaxRetry: maxRetries,
		}
		failCount = 0
		err := mockBroker.Publish(task)
		assert.Nil(t, err)

		time.Sleep(100 * time.Millisecond)
		cancel()
		wg.Wait()

		assert.Equal(t, maxRetries+1, failCount)
	})

	t.Run("should not execute task with no handler", func(t *testing.T) {
		mockBroker := NewMockBroker(1)
		wg := &sync.WaitGroup{}
		cfg := taskqueue.WorkerConfig{ID: 1, Broker: mockBroker, Backoff: nil, WG: wg}
		worker := taskqueue.DefaultWorkerFactory(cfg)
		worker.Register(taskName, taskHandler)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		wg.Add(1)
		go worker.Start(ctx)

		task := taskqueue.Task{
			Name: "invalid_task_name",
			Args: taskqueue.TaskArgs{},
		}
		taskHandlerCalled = false
		err := mockBroker.Publish(task)
		assert.Nil(t, err)

		time.Sleep(100 * time.Millisecond)
		cancel()
		wg.Wait()

		assert.False(t, taskHandlerCalled)
	})

	t.Run("should log and return after error", func(t *testing.T) {
		mockBroker := NewMockBroker(1)
		mockBroker.badConsume = true
		wg := &sync.WaitGroup{}
		cfg := taskqueue.WorkerConfig{ID: 1, Broker: mockBroker, Backoff: nil, WG: wg}
		worker := taskqueue.DefaultWorkerFactory(cfg)
		worker.Register(taskName, taskHandler)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var logOutput bytes.Buffer
		log.SetOutput(&logOutput)
		defer log.SetOutput(nil)

		wg.Add(1)
		go worker.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		assert.Contains(t, logOutput.String(), "Failed to consume tasks: simulated error")
	})

	t.Run("should log when task channel is closed", func(t *testing.T) {
		mockBroker := NewMockBroker(1)
		mockBroker.closeChannel = true
		wg := &sync.WaitGroup{}
		cfg := taskqueue.WorkerConfig{ID: 1, Broker: mockBroker, Backoff: nil, WG: wg}
		worker := taskqueue.DefaultWorkerFactory(cfg)
		worker.Register(taskName, taskHandler)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var logOutput bytes.Buffer
		log.SetOutput(&logOutput)
		defer log.SetOutput(nil)

		wg.Add(1)
		go worker.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		assert.Contains(t, logOutput.String(), "Task channel closed, worker exiting...")
	})
}
