package workers

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/store"
)

type TaskProcessor interface {
	ProcessTaskSendMail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskServerProcessor struct {
	server *asynq.Server
	store  store.Mongo
}

func NewTaskServerProcessor(opts asynq.RedisClientOpt, store store.Mongo) TaskProcessor {
	server := asynq.NewServer(opts, asynq.Config{})

	return &RedisTaskServerProcessor{
		server: server,
		store:  store,
	}
}
