package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const SERVER_ADDRES = ":3000"

func main() {
	client, err := Connect(context.Background(), "mongodb+srv://elisilas:i2CRksrspvydFl3S@cluster0.57zzc5l.mongodb.net/?retryWrites=true&w=majority")
	if err != nil {
		log.Print(err)
	}
  defer func() {
    if err := client.Disconnect(context.Background()); err != nil {
      log.Panic(err)
    }
  }()

	storage := NewStore(client)
	router := mux.NewRouter()
  postRouter := router.Methods(http.MethodPost).Subrouter()
  postRouter.HandleFunc("/coffee", handleFuncDecorator(storage.CreateCoffeeHandler))

	go func() {
		err := http.ListenAndServe(SERVER_ADDRES, router)
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
		Addr:         SERVER_ADDRES,
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
