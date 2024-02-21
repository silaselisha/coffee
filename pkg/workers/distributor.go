package workers

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface{
	SendEmail(ctx context.Context, payload PayloadSendMail, opts ...asynq.Option) error
}

type RedisTaskDistributor struct {
	clinet *asynq.Client
}

func NewRedisTaskDistributor(opts asynq.RedisClientOpt) TaskDistributor{
	client := asynq.NewClient(opts)
	return &RedisTaskDistributor{
		clinet: client,
	}
}

