package handler

import (
	"context"
	"log"

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

	tkn := token.NewToken(envs.SECRET_ACCESS_KEY)
	store := store.NewMongoClient(mongoClient)
	server.Store = store
	server.envs = envs
	server.token = tkn
	server.distributor = distributor

	validate := validator.New(validator.WithRequiredStructEnabled())
	server.vd = validate

	router := mux.NewRouter()
	productRoutes(router, server)
	userRoutes(router, server)

	server.Router = router
	return server
}
