package handler__test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/internal"
	"github.com/silaselisha/coffee-api/pkg/handler"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/workers"
	"go.mongodb.org/mongo-driver/mongo"
)

var mongoClient *mongo.Client
var product store.Item
var userID string
var productID string
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
	url := "/api/v1/login"
	body := map[string]interface{}{
		"email":    "admin@aws.ac.uk",
		"password": "Abstract$87",
	}

	userCred, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(userCred))
	recorder := httptest.NewRecorder()

	querier := handler.NewServer(context.Background(), mongoClient, distributor)
	server, ok := querier.(*handler.Server)
	if !ok {
		log.Fatal("Failed to initialize server")
	}

	server.Router.ServeHTTP(recorder, request)
	data, err := io.ReadAll(recorder.Body)
	if err != nil {
		log.Fatal(err)
	}

	var res struct{
		Status string `json:"status"`
		Token  string `json:"token"`
	}
	if err := json.Unmarshal(data, &res); err != nil {
		log.Fatal(err)
	}
	adminTestToken = res.Token

	redisOpts := asynq.RedisClientOpt{
		Addr: envs.REDIS_SERVER_ADDRESS,
	}
	distributor = workers.NewTaskClientDistributor(redisOpts)

	product = internal.CreateNewProduct()
	os.Exit(m.Run())
}
