package handler

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/token"
	"github.com/silaselisha/coffee-api/pkg/util"
	"github.com/silaselisha/coffee-api/pkg/workers"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	Router      *mux.Router
	Store       store.Mongo
	S3Client   	*util.CoffeeShopBucket
	vd          *validator.Validate
	envs        *util.Config
	token       token.Token
	distributor workers.TaskDistributor
}

func NewServer(ctx context.Context, mongoClient *mongo.Client, distributor workers.TaskDistributor) store.Querier {
	server := &Server{}

	envs, err := util.LoadEnvs("./../../")
	if err != nil {
		log.Panic(err)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Panic(err)
	}

	coffeShopBucket := util.NewS3Client(cfg, func(o *s3.Options) {
		o.Region = "us-east-1"
	})
	server.S3Client = coffeShopBucket

	tkn := token.NewToken(envs.SECRET_ACCESS_KEY)
	store := store.NewMongoClient(mongoClient)
	server.Store = store
	server.envs = envs
	server.token = tkn
	server.distributor = distributor

	validate := validator.New(validator.WithRequiredStructEnabled())
	server.vd = validate

	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	productRoutes(apiRouter, server)
	userRoutes(apiRouter, server)

	server.Router = router
	return server
}
