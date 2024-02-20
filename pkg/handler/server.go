package handler

import (
	"context"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
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
	mailer util.Transporter
}

func NewServer(ctx context.Context, mongoClient *mongo.Client) store.Querier {
	server := &Server{}

	envs, err := util.LoadEnvs("./../../")
	if err != nil {
		log.Panic(err)
	}

	transporter := util.NewSMTPTransporter(envs)

	tkn := token.NewToken(envs.SECRET_ACCESS_KEY)
	store := store.NewMongoClient(mongoClient)
	server.db = store
	server.envs = envs
	server.token = tkn
	server.mailer = transporter

	validate := validator.New(validator.WithRequiredStructEnabled())
	server.vd = validate

	router := mux.NewRouter()
	productRoutes(router, server)
	userRoutes(router, server)

	server.Router = router
	return server
}
