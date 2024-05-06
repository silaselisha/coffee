package api

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/internal"
	"github.com/silaselisha/coffee-api/internal/aws"
	"github.com/silaselisha/coffee-api/pkg/handler"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/token"
	"github.com/silaselisha/coffee-api/types"
	"github.com/silaselisha/coffee-api/workers"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	Router      *mux.Router
	Store       store.Mongo
	S3Client    *aws.CoffeeShopBucket
	vd          *validator.Validate
	envs        *types.Config
	token       token.Token
	distributor workers.TaskDistributor
}

func NewServer(ctx context.Context, mongoClient *mongo.Client, distributor workers.TaskDistributor, serveStaticFiles func() http.Handler) store.Querier {
	server := &Server{}

	envs, err := internal.LoadEnvs("./../../")
	if err != nil {
		log.Panic(err)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Panic(err)
	}

	coffeShopBucket := aws.NewS3Client(cfg, func(o *s3.Options) {
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

	templ := handler.NewTemplate()
	render(router, templ, serveStaticFiles) // serve static files

	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	productRoutes(apiRouter, server)
	userRoutes(apiRouter, server)

	server.Router = router
	return server
}

func render(router *mux.Router, templ handler.Querier, serveStaticFiles func() http.Handler) {
	router.PathPrefix("/public/").Handler(serveStaticFiles())
	router.HandleFunc("/", internal.HandleFuncDecorator(templ.RenderHomePageHandler))
}
