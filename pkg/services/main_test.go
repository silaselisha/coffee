package services

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
var id string

func TestMain(m *testing.M) {
	fmt.Println("RUNNING")
	var err error
	envs, err := util.LoadEnvs("./../..")
	if err != nil {
		log.Fatal(err)
	}

	mongoClient, err = util.Connect(context.Background(), envs.DBUri)
	if err != nil {
		log.Fatal(err)
	}

	product = util.CreateNewProduct()
	os.Exit(m.Run())
}
