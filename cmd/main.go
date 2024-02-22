package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/handler"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"github.com/silaselisha/coffee-api/pkg/workers"
	"github.com/sirupsen/logrus"
)

func main() {
	config, err := util.LoadEnvs("./..")
	if err != nil {
		log.Panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	mongo_client, err := util.Connect(ctx, config.DB_URI)
	if err != nil {
		log.Panic(err)
	}

	defer func() {
		if err := mongo_client.Disconnect(ctx); err != nil {
			log.Panic(err)
		}
	}()

	redisOpts := asynq.RedisClientOpt{
		Addr: config.REDIS_SERVER_ADDRESS,
	}

	distributor := workers.NewTaskClientDistributor(redisOpts)
	querier := handler.NewServer(ctx, mongo_client, distributor)
	server := querier.(*handler.Server)
	go taskProcessor(redisOpts, server.Store)

	err = http.ListenAndServe(config.SERVER_ADDRESS, server.Router)
	if err != nil {
		logrus.Fatal(err)
	}
}

func taskProcessor(opts asynq.RedisClientOpt, store store.Mongo) {
	processor := workers.NewTaskServerProcessor(opts, store)
	fmt.Printf("worker process on go: @%v\n", time.Now())
	err := processor.Start()
	if err != nil {
		log.Fatal(err)
	}
}
