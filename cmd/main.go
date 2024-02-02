package main

import (
	"context"
	"log"
	"net/http"

	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/api"
	"github.com/silaselisha/coffee-api/pkg/util"
	"github.com/sirupsen/logrus"
)

func main() {
	config, err := util.LoadEnvs("./..")
	if err != nil {
		log.Print(err)
	}

	client, err := util.Connect(context.Background(), config.DBUri)
	if err != nil {
		log.Print(err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Panic(err)
		}
	}()

	storage := store.NewMongoClient(client)
	server := api.NewServer(storage)

	err = http.ListenAndServe(config.ServerAddrs, server.Router)
	if err != nil {
		logrus.Fatal(err)
	}
}
