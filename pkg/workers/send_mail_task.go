package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const SEND_VERIFICATION_EMAIL string = "task:send_verification_email"
const SEND_PASSWORD_RESET_EMAIL string = "task:send_password_reset_email"

func (distributor *RedisTaskClientDistributor) SendVerificationMailTask(ctx context.Context, payload *util.PayloadSendMail, opts ...asynq.Option) error {
	payloadBuffer, err := json.Marshal(payload)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(SEND_VERIFICATION_EMAIL, payloadBuffer, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Printf("Enqueued task: %v of max retries: %v on payload: %v\n", info.Type, info.MaxRetry, string(info.Payload))
	return nil
}

func (distributor *RedisTaskClientDistributor) SendPasswordResetMailTask(ctx context.Context, payload *util.PayloadSendMail, opts ...asynq.Option) error {
	payloadBuffer, err := json.Marshal(payload)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(SEND_PASSWORD_RESET_EMAIL, payloadBuffer, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Printf("Enqueued task: %v of max retries: %v on payload: %v\n", info.Type, info.MaxRetry, string(info.Payload))
	return nil
}

func (processor *RedisTaskServerProcessor) ProcessTaskSendVerificationMail(ctx context.Context, task *asynq.Task) error {
	user, err := getUserByEmail(ctx, processor, task)
	if err != nil {
		return fmt.Errorf("error occured while retreiving user %w", err)
	}

	transporter := util.NewSMTPTransporter(&processor.envs)
	message := fmt.Sprintf("http://localhost:3000/verify?token=%s&timestamp=%d", user.Id.Hex(), util.ResetToken(2880))

	err = transporter.MailSender(ctx, user.Email, []byte(message))
	if err != nil {
		return fmt.Errorf("error occured while sending a verification mail to %s at %v err %w", user.Email, time.Now(), err)
	}

	fmt.Printf("processing %s at %v\n", task.Type(), time.Now())
	return nil
}

func (processor *RedisTaskServerProcessor) ProcessTaskSendResetPasswordMail(ctx context.Context, task *asynq.Task) error {
	user, err := getUserByEmail(ctx, processor, task)
	if err != nil {
		return fmt.Errorf("error occured while retreiving user %w", err)
	}

	transporter := util.NewSMTPTransporter(&processor.envs)
	message := fmt.Sprintf("http://localhost:3000/resetpassword?token=%s&timestamp=%d", user.Id.Hex(), util.ResetToken(2880))

	err = transporter.MailSender(ctx, user.Email, []byte(message))
	if err != nil {
		return fmt.Errorf("error occured while sending a verification mail to %s at %v err %w", user.Email, time.Now(), err)
	}

	fmt.Printf("processing %s at %v\n", task.Type(), time.Now())
	return nil
}

func getUserByEmail(ctx context.Context, processor *RedisTaskServerProcessor, task *asynq.Task) (store.User, error) {
	var Payload util.PayloadSendMail
	err := json.Unmarshal(task.Payload(), &Payload)
	if err != nil {
		fmt.Print(time.Now())
		return store.User{}, fmt.Errorf("unmarshalling error %w", err)
	}

	var user store.User
	users := processor.store.Collection(ctx, "coffeeshop", "users")
	curr := users.FindOne(ctx, bson.D{{Key: "email", Value: Payload.Email}})
	err = curr.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Print(time.Now())
			return store.User{}, fmt.Errorf("document not found %w", err)
		}
		return store.User{}, fmt.Errorf("internal server error %w", err)
	}

	return user, nil
}

func (processor *RedisTaskServerProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(SEND_VERIFICATION_EMAIL, processor.ProcessTaskSendVerificationMail)
	mux.HandleFunc(SEND_PASSWORD_RESET_EMAIL, processor.ProcessTaskSendResetPasswordMail)
	mux.HandleFunc(UPLOAD_S3_OBJECT, processor.ProcessTaskUploadS3Object)
	mux.HandleFunc(UPLOAD_MULTIPLE_S3_OBJECTS, processor.ProcessTaskMultipleUploadS3Object)
	mux.HandleFunc(DELETE_S3_OBJECT, processor.ProcessTaskDeleteS3Object)

	return processor.server.Start(mux)
}
