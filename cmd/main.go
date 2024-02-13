package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/silaselisha/coffee-api/pkg/handler"
	"github.com/silaselisha/coffee-api/pkg/util"
	"github.com/sirupsen/logrus"
)

func main() {
	config, err := util.LoadEnvs("./..")
	if err != nil {
		log.Panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20 * time.Second)
	defer cancel()
	
	client, err := util.Connect(ctx, config.DBUri)
	if err != nil {
		log.Panic(err)
	}

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Panic(err)
		}
	}()
	server := handler.NewServer(ctx, client)
	router := server.(*handler.Server)

	err = http.ListenAndServe(config.ServerAddrs, router.Router)
	if err != nil {
		logrus.Fatal(err)
	}
}
