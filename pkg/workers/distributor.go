package workers

import (
	"context"

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

func NewTaskClientDistributor(opts asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(opts)
	return &RedisTaskClientDistributor{
		client: client,
	}
}
