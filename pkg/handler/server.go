package handler

import (
	"context"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/token"
	"github.com/silaselisha/coffee-api/pkg/util"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	Router *mux.Router
	db     store.Mongo
	vd     *validator.Validate
	envs   *util.Config
	token  token.Token
	RDB    *redis.Client
}

func NewServer(ctx context.Context, mongoClient *mongo.Client) store.Querier {
	server := &Server{}

	envs, err := util.LoadEnvs("./../../")
	if err != nil {
		log.Panic(err)
	}

	tkn := token.NewToken(envs.SecretAccessKey)
	store := store.NewMongoClient(mongoClient)
	server.db = store
	server.envs = envs
	server.token = tkn

	validate := validator.New(validator.WithRequiredStructEnabled())
	server.vd = validate

	router := mux.NewRouter()
	productRoutes(router, server)
	userRoutes(router, server)

	server.Router = router
	return server
}
