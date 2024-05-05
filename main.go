package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/handler"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/types"
	"github.com/silaselisha/coffee-api/internal"
	"github.com/silaselisha/coffee-api/internal/aws"
	"github.com/silaselisha/coffee-api/workers"
)

func main() {
	envs, err := internal.LoadEnvs(".")
	if err != nil {
		log.Panic(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	mongo_client, err := internal.Connect(ctx, envs)
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
	querier := handler.NewServer(ctx, mongo_client, distributor, public)
	server := querier.(*handler.Server)

	cfg, err := config.LoadDefaultConfig(ctx, func(lo *config.LoadOptions) error { return nil })
	if err != nil {
		log.Panic(err)
		return
	}
	client := aws.NewS3Client(cfg, func(o *s3.Options) {
		o.Region = "us-east-1"
	})

	go taskProcessor(redisOpts, server.Store, *envs, client)

	fmt.Printf("serving HTTP/REST server\n")
	fmt.Printf("http://localhost:%v/\n", envs.SERVER_REST_ADDRESS)

	err = http.ListenAndServe(envs.SERVER_REST_ADDRESS, server.Router)
	if err != nil {
		log.Panic(err)
		return
	}
}

func taskProcessor(opts asynq.RedisClientOpt, store store.Mongo, envs types.Config, client *aws.CoffeeShopBucket) {
	processor := workers.NewTaskServerProcessor(opts, store, envs, client)
	log.Print("worker process on")
	err := processor.Start()
	if err != nil {
		log.Panic(err)
		return
	}
}
