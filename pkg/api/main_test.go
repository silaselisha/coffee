package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
)

var testMonogoStore store.Mongo
var product store.Item
var id string

func TestMain(m *testing.M) {
	fmt.Println("RUNNING")
	var err error
	envs, err := util.LoadEnvs("./../..")
	if err != nil {
		log.Fatal(err)
	}

	client, err := util.Connect(context.Background(), envs.DBUri)
	if err != nil {
		log.Fatal(err)
	}

	testMonogoStore = store.NewMongoClient(client)
	product = util.CreateNewProduct()
	os.Exit(m.Run())
}
