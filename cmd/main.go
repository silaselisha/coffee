package main

import (
	"context"
	"fmt"
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

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	fmt.Println(config.DB_URI)
	fmt.Println(config.SERVER_ADDRESS)
	fmt.Println("http://localhost:3000")
	mongo_client, err := util.Connect(ctx, config.DB_URI)
	if err != nil {
		log.Panic(err)
	}

	defer func() {
		if err := mongo_client.Disconnect(ctx); err != nil {
			log.Panic(err)
		}
	}()

	querier := handler.NewServer(ctx, mongo_client)
	server := querier.(*handler.Server)

	err = http.ListenAndServe(config.SERVER_ADDRESS, server.Router)
	if err != nil {
		logrus.Fatal(err)
	}
}
