package main

import (
	"context"
	"net/http"
	"time"

	"log"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/handler"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"github.com/silaselisha/coffee-api/pkg/workers"
)

func main() {
	config, err := util.LoadEnvs("./..")
	if err != nil {
		log.Panic(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	mongo_client, err := util.Connect(ctx, config.DB_URI)
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
		Addr: config.REDIS_SERVER_ADDRESS,
	}

	distributor := workers.NewTaskClientDistributor(redisOpts)
	querier := handler.NewServer(ctx, mongo_client, distributor)
	server := querier.(*handler.Server)
	go taskProcessor(redisOpts, server.Store)

	err = http.ListenAndServe(config.SERVER_ADDRESS, server.Router)
	if err != nil {
		log.Panic()
		return
	}
}

func taskProcessor(opts asynq.RedisClientOpt, store store.Mongo) {
	processor := workers.NewTaskServerProcessor(opts, store)
	log.Print("worker process on")
	err := processor.Start()
	if err != nil {
		log.Panic(err)
		return
	}
}
