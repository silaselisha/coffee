package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/util"
)

var testMonogoStore store.Mongo
var product Item
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
	product = createNewProduct()
	os.Exit(m.Run())
}
