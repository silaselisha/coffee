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
	"github.com/rs/cors"
	"github.com/silaselisha/coffee-api/internal"
	"github.com/silaselisha/coffee-api/internal/aws"
	"github.com/silaselisha/coffee-api/pkg/api"
	"github.com/silaselisha/coffee-api/pkg/handler"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/types"
	"github.com/silaselisha/coffee-api/workers"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	envs, err := internal.LoadEnvs(".")
	if err != nil {
		log.Panic(err)
		return
	}

	server, redisOpts,  mongo_client, coffeeShopS3Bucket := mainHelper(ctx, envs)
	defer func() {
		if err := mongo_client.Disconnect(ctx); err != nil {
			log.Panic(err)
			return
		}
	}()

	go taskProcessor(redisOpts, server.Store, *envs, coffeeShopS3Bucket)

	fmt.Printf("serving HTTP/REST server\n")
	fmt.Printf("http://localhost:%v/\n", envs.SERVER_REST_ADDRESS)

	handler := cors.Default().Handler(server.Router)
	err = http.ListenAndServe(envs.SERVER_REST_ADDRESS, handler)
	if err != nil {
		log.Panic(err)
		return
	}
}

func mainHelper(ctx context.Context, envs *types.Config) (server *api.Server, redisOpts asynq.RedisClientOpt, mongo_client *mongo.Client, coffeeShopS3Bucket aws.CoffeeShopBucket) {
	mongo_client, err := internal.Connect(ctx, envs)
	if err != nil {
		log.Panic(err)
		return
	}

	redisOpts = asynq.RedisClientOpt{
		Addr: envs.REDIS_SERVER_ADDRESS,
	}

	templQueries := handler.NewTemplate(".")

	distributor := workers.NewTaskClientDistributor(redisOpts)
	querier := api.NewServer(ctx, envs, mongo_client, distributor, templQueries, public)
	server = querier.(*api.Server)

	cfg, err := config.LoadDefaultConfig(ctx, func(lo *config.LoadOptions) error { return nil })
	if err != nil {
		log.Panic(err)
		return
	}

	coffeeShopS3Bucket = aws.NewS3Client(cfg, func(o *s3.Options) {
		o.Region = "us-east-1"
	})
	return
}

func taskProcessor(opts asynq.RedisClientOpt, store store.Mongo, envs types.Config, coffeeShopS3Bucket aws.CoffeeShopBucket) {
	processor := workers.NewTaskServerProcessor(opts, store, envs, coffeeShopS3Bucket)
	log.Print("worker process on")
	err := processor.Start()
	if err != nil {
		log.Panic(err)
		return
	}
}
