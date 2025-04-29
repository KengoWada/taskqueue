package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KengoWada/taskqueue"
	"github.com/KengoWada/taskqueue/brokers/redisbroker"
)

const sendEmailTaskName string = "send_email"

func sendEmailTask(args taskqueue.TaskArgs) error {
	email := args["email"].(string)

	time.Sleep(5 * time.Second) // Simulated work
	if email == "fail@example.com" {
		return fmt.Errorf("simulated failure")
	}

	fmt.Printf("Sending email to %s\n", email)
	return nil
}

func main() {
	broker, err := redisbroker.NewRedisBroker("localhost:6379", "app_tasks")
	if err != nil {
		fmt.Println(err)
		return
	}

	backoffPolicy := &taskqueue.BackoffPolicy{
		BaseDelay: 2 * time.Second,
		MaxDelay:  60 * time.Second,
	}

	manager := taskqueue.NewManager(broker, taskqueue.DefaultWorkerFactory, 5, taskqueue.WithBackoffPolicy(backoffPolicy))
	manager.RegisterTask(sendEmailTaskName, sendEmailTask)
	manager.Start()

	manager.PublishTask(sendEmailTaskName, taskqueue.TaskArgs{"email": "kengo@cia.gov"}, 3)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	manager.Stop()
}
