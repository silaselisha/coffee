package workers

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/store"
)

const (
	DefaultQueue  = "default"
	CriticalQueue = "critical"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendMail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskServerProcessor struct {
	server *asynq.Server
	store  store.Mongo
}

func NewTaskServerProcessor(opts asynq.RedisClientOpt, store store.Mongo) TaskProcessor {
	server := asynq.NewServer(opts, asynq.Config{
		Queues: map[string]int{CriticalQueue: 1, DefaultQueue: 2},
	})

	return &RedisTaskServerProcessor{
		server: server,
		store:  store,
	}
}
