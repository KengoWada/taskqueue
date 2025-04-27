package redisbroker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/KengoWada/taskqueue"
	"github.com/redis/go-redis/v9"
)

type RedisBroker struct {
	client *redis.Client
	queue  string
}

// NewRedisBroker creates a new RedisBroker instance.
//
// It connects to the specified Redis server at the given address and sets up a queue name
// that will be used for publishing and consuming tasks.
//
// Parameters:
//   - addr: Redis server address, e.g., "localhost:6379".
//   - queue: Name of the Redis list (queue) to use for tasks.
//
// Returns:
//   - A pointer to a RedisBroker ready to publish and consume tasks.
//
// Example usage:
//
//	broker := NewRedisBroker("localhost:6379", "taskqueue")
func NewRedisBroker(addr, queue string) *RedisBroker {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RedisBroker{
		client: client,
		queue:  queue,
	}
}

// Publish pushes the given task onto the Redis queue.
// The task is marshaled into JSON format before being stored.
func (r *RedisBroker) Publish(task taskqueue.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return r.client.LPush(context.Background(), r.queue, data).Err()
}

// Consume continuously listens for tasks from the Redis queue.
//
// Internally, it uses BRPop with a blocking timeout to wait for new tasks.
// If an error occurs during BRPop or unmarshaling, it simply continues listening without terminating.
//
// The returned channel will emit tasks as they are received.
func (r *RedisBroker) Consume() (<-chan taskqueue.Task, error) {
	out := make(chan taskqueue.Task)
	go func() {
		for {
			result, err := r.client.BRPop(context.Background(), 0*time.Second, r.queue).Result()
			if err != nil {
				continue
			}
			var task taskqueue.Task
			json.Unmarshal([]byte(result[1]), &task)
			out <- task
		}
	}()
	return out, nil
}
