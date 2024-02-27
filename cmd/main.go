package main

import (
	"context"
	"net/http"
	"time"

	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/handler"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"github.com/silaselisha/coffee-api/pkg/workers"
)

func main() {
	envs, err := util.LoadEnvs("./..")
	if err != nil {
		log.Panic(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	mongo_client, err := util.Connect(ctx, envs.DB_URI)
	if err != nil {
		log.Panic(err)
		return
	}

	defer func() {
		if err := mongo_client.Disconnect(ctx); err != nil {
			log.Panic(err)
			return
		}
	}()

	redisOpts := asynq.RedisClientOpt{
		Addr: envs.REDIS_SERVER_ADDRESS,
	}

	distributor := workers.NewTaskClientDistributor(redisOpts)
	querier := handler.NewServer(ctx, mongo_client, distributor)
	server := querier.(*handler.Server)

	cfg, err := config.LoadDefaultConfig(ctx, func(lo *config.LoadOptions) error { return nil })
	if err != nil {
		log.Panic(err)
		return
	}
	client := util.NewS3Client(cfg, func(o *s3.Options) {
		o.Region = "us-east-1"
	})

	go taskProcessor(redisOpts, server.Store, *envs, client)

	err = http.ListenAndServe(envs.SERVER_ADDRESS, server.Router)
	if err != nil {
		log.Panic()
		return
	}
}

func taskProcessor(opts asynq.RedisClientOpt, store store.Mongo, envs util.Config, client *util.CoffeeShopBucket) {
	processor := workers.NewTaskServerProcessor(opts, store, envs, client)
	log.Print("worker process on")
	err := processor.Start()
	if err != nil {
		log.Panic(err)
		return
	}
}
