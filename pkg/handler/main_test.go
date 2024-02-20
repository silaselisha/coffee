package handler

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"go.mongodb.org/mongo-driver/mongo"
)

var mongoClient *mongo.Client
var product store.Item
var productID string
var userID string
var userTestToken string
var adminID string
var adminTestToken string

func TestMain(m *testing.M) {
	fmt.Println("RUNNING")
	var err error
	envs, err := util.LoadEnvs("./../..")
	if err != nil {
		log.Fatal(err)
	}

	mongoClient, err = util.Connect(context.Background(), envs.DB_URI)
	if err != nil {
		log.Fatal(err)
	}
	adminID = "65d24b2df041357fe87113bc"

	product = util.CreateNewProduct()
	os.Exit(m.Run())
}
