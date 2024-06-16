package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/types"
)

const (
	UPLOAD_S3_OBJECT           = "task:upload_s3_object"
	UPLOAD_MULTIPLE_S3_OBJECTS = "task:upload_multiple_s3_objects"
	DELETE_S3_OBJECT           = "task:delete_s3_object"
	SEND_VERIFICATION_EMAIL    = "task:send_verification_email"
	SEND_PASSWORD_RESET_EMAIL  = "task:send_password_reset_email"
)

type TaskDistributor interface {
	VerificationMailTask(ctx context.Context, payload *types.PayloadSendMail, opts ...asynq.Option) error
	PasswordResetMailTask(ctx context.Context, payload *types.PayloadSendMail, opts ...asynq.Option) error
	S3ObjectUploadTask(ctx context.Context, payload *types.PayloadUploadImage, opts ...asynq.Option) error
	MultipleS3ObjectUploadTask(ctx context.Context, payload []*types.PayloadUploadImage, opts ...asynq.Option) error
	S3ObjectDeleteTask(ctx context.Context, images []string, opts ...asynq.Option) error
}

type RedisClientTaskDistributor struct {
	client *asynq.Client
}

func NewTaskClientDistributor(opts asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(opts)
	return &RedisClientTaskDistributor{
		client: client,
	}
}

func (dist *RedisClientTaskDistributor) VerificationMailTask(ctx context.Context, payload *types.PayloadSendMail, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(SEND_VERIFICATION_EMAIL, data, opts...)
	info, err := dist.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Printf("Enqueued task: %v of max retries: %v on payload: %v\n", info.Type, info.MaxRetry, string(info.Payload))
	return nil
}

func (dist *RedisClientTaskDistributor) PasswordResetMailTask(ctx context.Context, payload *types.PayloadSendMail, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(SEND_PASSWORD_RESET_EMAIL, data, opts...)
	info, err := dist.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Printf("Enqueued task: %v of max retries: %v on payload: %v\n", info.Type, info.MaxRetry, string(info.Payload))
	return nil
}

func (dist *RedisClientTaskDistributor) S3ObjectUploadTask(ctx context.Context, payload *types.PayloadUploadImage, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(UPLOAD_S3_OBJECT, data, opts...)
	info, err := dist.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Println("send s3 object task")
	fmt.Printf("Enqueued task: %v of max retries: %v\n", info.Type, info.MaxRetry)
	return nil
}

func (dist *RedisClientTaskDistributor) MultipleS3ObjectUploadTask(ctx context.Context, payload []*types.PayloadUploadImage, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error %w", err)
	}

	task := asynq.NewTask(UPLOAD_MULTIPLE_S3_OBJECTS, data, opts...)
	info, err := dist.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Printf("Enqueued task: %v of max retries: %v\n", info.Type, info.MaxRetry)
	return nil
}

func (dist *RedisClientTaskDistributor) S3ObjectDeleteTask(ctx context.Context, images []string, opts ...asynq.Option) error {
	data, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("error occured while marshaling %w", err)
	}

	task := asynq.NewTask(DELETE_S3_OBJECT, data, opts...)
	info, err := dist.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("enqueueing task error %w", err)
	}

	fmt.Printf("Enqueued task: %v of max retries: %v\n", info.Type, info.MaxRetry)
	return nil
}
