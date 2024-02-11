package main

import (
	"context"
	"log"
	"net/http"

	"github.com/silaselisha/coffee-api/pkg/handler"
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

	server := handler.NewServer(client)
	router, ok := server.(*handler.Server)
	if !ok {
		logrus.Error("internal server error")
	}

	err = http.ListenAndServe(config.ServerAddrs, router.Router)
	if err != nil {
		logrus.Fatal(err)
	}
}
