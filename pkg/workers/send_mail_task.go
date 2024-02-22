package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PayloadSendMail struct {
	Email string `json:"email"`
}

const SEND_MAIL_TASK_KEY string = "task:send_mail_key"

func (distributor *RedisTaskClientDistributor) SendMailTask(ctx context.Context, payload *PayloadSendMail, opts ...asynq.Option) error {
	payloadBuffer, err := json.Marshal(payload)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(SEND_MAIL_TASK_KEY, payloadBuffer, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Printf("Enqueued task: %v of max retries: %v on payload: %v\n", info.Type, info.MaxRetry, string(info.Payload))
	return nil
}

func (processor *RedisTaskServerProcessor) ProcessTaskSendMail(ctx context.Context, task *asynq.Task) error {
	var Payload PayloadSendMail
	err := json.Unmarshal(task.Payload(), &Payload)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("unmarshalling error %w", err)
	}

	var user store.User
	users := processor.store.Collection(ctx, "coffeeshop", "users")
	curr := users.FindOne(ctx, bson.D{{Key: "email", Value: Payload.Email}})
	err = curr.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Print(time.Now())
			return fmt.Errorf("document not found %w", err)
		}
		return fmt.Errorf("internal server error %w", err)
	}

	fmt.Printf("processing %s at %v\n", task.Type(), time.Now())
	return nil
}

func (processor *RedisTaskServerProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(SEND_MAIL_TASK_KEY, processor.ProcessTaskSendMail)
	
	return processor.server.Start(mux)
}