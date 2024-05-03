package handler__test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/internal"
	"github.com/silaselisha/coffee-api/workers"
	"go.mongodb.org/mongo-driver/mongo"
)

var mongoClient *mongo.Client
var product store.Item
var userID string
var userTestToken string
var adminID string
var adminTestToken string
var distributor workers.TaskDistributor

func TestMain(m *testing.M) {
	fmt.Println("RUNNING")
	var err error
	envs, err := internal.LoadEnvs("./../../..")
	if err != nil {
		log.Fatal(err)
	}

	mongoClient, err = internal.Connect(context.Background(), envs)
	if err != nil {
		log.Fatal(err)
	}
	adminID = "66348187510f523cea4fbd7a"

	redisOpts := asynq.RedisClientOpt{
		Addr: envs.REDIS_SERVER_ADDRESS,
	}
	distributor = workers.NewTaskClientDistributor(redisOpts)

	product = internal.CreateNewProduct()
	os.Exit(m.Run())
}
