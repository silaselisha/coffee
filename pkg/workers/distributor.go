package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/util"
)

type TaskDistributor interface {
	SendVerificationMailTask(ctx context.Context, payload *util.PayloadSendMail, opts ...asynq.Option) error
	SendPasswordResetMailTask(ctx context.Context, payload *util.PayloadSendMail, opts ...asynq.Option) error
	SendS3ObjectUploadTask(ctx context.Context, payload *util.PayloadUploadImage, opts ...asynq.Option) error
	SendMultipleS3ObjectUploadTask(ctx context.Context, payload []*util.PayloadUploadImage, opts ...asynq.Option) error
	SendS3ObjectDeleteTask(ctx context.Context, images []string, opts ...asynq.Option) error
}

type RedisTaskClientDistributor struct {
	client *asynq.Client
}

const (
	UPLOAD_S3_OBJECT           string = "task:upload_s3_object"
	UPLOAD_MULTIPLE_S3_OBJECTS string = "task:upload_multiple_s3_objects"
	DELETE_S3_OBJECT           string = "task:delete_s3_object"
	SEND_VERIFICATION_EMAIL    string = "task:send_verification_email"
	SEND_PASSWORD_RESET_EMAIL  string = "task:send_password_reset_email"
)

func NewTaskClientDistributor(opts asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(opts)
	return &RedisTaskClientDistributor{
		client: client,
	}
}

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

func (distributor *RedisTaskClientDistributor) SendS3ObjectUploadTask(ctx context.Context, payload *util.PayloadUploadImage, opts ...asynq.Option) error {
	payloadBuffer, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(UPLOAD_S3_OBJECT, payloadBuffer, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Println("send s3 object task")
	fmt.Printf("Enqueued task: %v of max retries: %v\n", info.Type, info.MaxRetry)
	return nil
}

func (distributor *RedisTaskClientDistributor) SendMultipleS3ObjectUploadTask(ctx context.Context, payload []*util.PayloadUploadImage, opts ...asynq.Option) error {
	payloadBuffer, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(UPLOAD_MULTIPLE_S3_OBJECTS, payloadBuffer, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		fmt.Print(time.Now())
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Printf("Enqueued task: %v of max retries: %v\n", info.Type, info.MaxRetry)
	return nil
}

func (distributor *RedisTaskClientDistributor) SendS3ObjectDeleteTask(ctx context.Context, images []string, opts ...asynq.Option) error {
	payload, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("error occured while marshaling %w", err)
	}

	task := asynq.NewTask(DELETE_S3_OBJECT, payload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Printf("Enqueued task: %v of max retries: %v\n", info.Type, info.MaxRetry)
	return nil
}
