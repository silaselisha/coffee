package workers

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
)

const (
	DefaultQueue  = "default"
	CriticalQueue = "critical"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerificationMail(ctx context.Context, task *asynq.Task) error
	ProcessTaskUploadS3Object(ctx context.Context, task *asynq.Task) error
}

type RedisTaskServerProcessor struct {
	server *asynq.Server
	store  store.Mongo
	envs   util.Config
	client *util.CoffeeShopBucket
}

func NewTaskServerProcessor(opts asynq.RedisClientOpt, store store.Mongo, envs util.Config, client *util.CoffeeShopBucket) TaskProcessor {

	server := asynq.NewServer(opts, asynq.Config{
		Queues: map[string]int{CriticalQueue: 1, DefaultQueue: 2},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Error().Err(err).Str("type", task.Type()).Msg("process failed")
		}),
	})

	return &RedisTaskServerProcessor{
		server: server,
		store:  store,
		envs:   envs,
		client: client,
	}
}
