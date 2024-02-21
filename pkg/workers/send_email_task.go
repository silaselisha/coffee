package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PayloadSendMail struct {
	Email string `json:"email"`
}

const SEND_EMAIL_KEY = "task:send_mail_key"

func (distributor *RedisTaskDistributor) SendEmail(ctx context.Context, pyld PayloadSendMail, opts ...asynq.Option) error {
	payload, err := json.Marshal(pyld)
	if err != nil {
		fmt.Print(err)
		return fmt.Errorf("payload marshling error...%w", err)
	}

	task := asynq.NewTask(SEND_EMAIL_KEY, payload, opts...)
	info, err := distributor.clinet.EnqueueContext(ctx, task)
	if err != nil {
		fmt.Print(err)
		return fmt.Errorf("task queueing error...%w", err)
	}

	fmt.Print(info)
	return nil
}

func (processor *RedisTaskProcessor) ProcessSendMailTask(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendMail
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		log.Print(err)
		return fmt.Errorf("payload unmarshal error %w", err)
	}

	users := processor.store.Collection(ctx, "coffeeshop", "users")
	var user store.User
	curr := users.FindOne(ctx, bson.D{{Key: "email", Value: payload.Email}})
	err = curr.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("user not found %w", err)
		}
		return fmt.Errorf("internal server error %w", err)
	}

	log.Print("task is being processed...")
	return nil
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(SEND_EMAIL_KEY, processor.ProcessSendMailTask)
	err := processor.server.Start(mux)
	return err
}