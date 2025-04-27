# TaskQueue: Go-based Distributed Task Queue

TaskQueue is a Go library that provides a simple and efficient way to manage and execute asynchronous tasks. It's inspired by [Celery](https://docs.celeryq.dev/en/stable/getting-started/introduction.html) and designed to be highly extensible, allowing you to easily distribute tasks across multiple workers.

## ⚠️ Warning: Not Production Ready

### Features

- **Task Definition**: Easily define tasks with customizable arguments and retry behavior.

- **Worker**: Workers fetch tasks from a broker and execute them with optional retry and backoff strategies.

- **Manager**: Manage multiple workers, task registration, and graceful shutdown handling.

- **Broker** Interface: Abstracts task transport; easily extensible to different backends.

- **Redis Broker**: Built-in Redis-based broker for task queueing.

- **Backoff Policy**: Optional exponential backoff with jitter for retrying failed tasks.

### Coming Soon

- **Scheduled Tasks**: Ability to schedule tasks to run at specific times or intervals.

- **Custom Logger Support**: Allow users to inject their own logging system.

- **Better Error Handling and Dead Letter Queues**: Capture and manage tasks that permanently fail after retries.

- **New Broker Implementations**: Add support for new brokers

  - **RabbitMQ**: Native support for RabbitMQ as a task transport backend.

  - **GCP Pub/Sub**: Experimental support for Google Cloud Pub/Sub as a broker option.

### Trying it out

Since the library is still under development and not yet published, you can clone the repository and run the examples locally:

```sh
git clone git@github.com:KengoWada/taskqueue.git
cd taskqueue/example
go run main.go
```

The example/ directory contains a simple usage demonstration to help you get started quickly.
