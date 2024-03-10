package gapi

import (
	"github.com/silaselisha/coffee-api/pkg/pb"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	pb.UnimplementedOrderServiceServer
	mongo store.Mongo
	envs  *util.Config
}

func NewServer(envs *util.Config, mongo *mongo.Client) *Server {
	store := store.NewMongoClient(mongo)
	return &Server{
		envs:  envs,
		mongo: store,
	}
}
