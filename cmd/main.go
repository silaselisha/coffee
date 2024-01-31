package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/pkg/products"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/util"
	"github.com/sirupsen/logrus"
)

var validate *validator.Validate

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

	validate = validator.New(validator.WithRequiredStructEnabled())
	
	storage := store.NewStore(client)
	products := products.NewProduct(storage)
	router := mux.NewRouter()
	postRouter := router.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/coffee", util.HandleFuncDecorator(products.CreateProductHandler))

	go func() {
		err := http.ListenAndServe(config.ServerAddrs, router)
		if err != nil {
			logrus.Fatal(err)
		}
	}()

	sig_chan := make(chan os.Signal, 4)
	signal.Notify(sig_chan, os.Interrupt)
	signal.Notify(sig_chan, syscall.SIGTERM)

	<-sig_chan
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	server := http.Server{
		Addr:         config.ServerAddrs,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		Handler:      router,
	}
	err = server.Shutdown(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
}
