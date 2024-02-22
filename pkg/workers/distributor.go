package workers

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface{
	SendMailTask(ctx context.Context, payload *PayloadSendMail, opts ...asynq.Option) error
}

type RedisTaskClientDistributor struct {
	client *asynq.Client
}

func NewTaskDistributor(opts asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(opts)
	return &RedisTaskClientDistributor{
		client: client,
	}
}