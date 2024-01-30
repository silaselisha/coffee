package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
  "github.com/silaselisha/coffee-api/store"
)

type ProductCoffeeI interface {
  CreateCoffeeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

type ProductCoffee struct {
	db store.Store
}

func NewProduct(storage store.Store) ProductCoffeeI {
  return &ProductCoffee {
    db: storage,
  }
}

func (s *ProductCoffee) CreateCoffeeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection, err := s.db.Collection(ctx, "coffeeshop", "coffee")
	if err != nil {
		w.Write([]byte("invalid operation"))
		return err
	}

	data := new(coffee)
	coffeeBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("invalid data operation"))
		return err
	}

  data.Id = primitive.NewObjectID()
  data.CreatedAt = time.Now()
  data.UpdatedAt = time.Now()

	err = json.Unmarshal(coffeeBytes, data)
	if err != nil {
		w.Write([]byte("invalid data operation"))
		return err
	}

	_, err = collection.InsertOne(ctx, data)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("invalid insertion operation"))
		return err
	}

  return nil
}
