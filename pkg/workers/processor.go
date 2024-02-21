package workers

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/store"
)

type TaskProcessor interface {
	Start() error
	ProcessSendMailTask(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  store.Mongo
}

func NewRedisTaskProcessor(opts asynq.RedisClientOpt, db store.Mongo) TaskProcessor {
	server := asynq.NewServer(opts, asynq.Config{})
	return &RedisTaskProcessor{
		server: server,
		store:  db,
	}
}
