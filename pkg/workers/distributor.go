package workers

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface{
	SendVerificationMailTask(ctx context.Context, payload *PayloadSendMail, opts ...asynq.Option) error
	SendPasswordResetMailTask(ctx context.Context, payload *PayloadSendMail, opts ...asynq.Option) error
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