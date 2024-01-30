package main

import (
	"context"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Store interface {
	Collection(ctx context.Context, db string, collection string) (*mongo.Collection, error)
	CreateCoffeeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

type MongoStore struct {
	client *mongo.Client
}

func NewStore(client *mongo.Client) Store {
	return &MongoStore{
		client: client,
	}
}

func (ms *MongoStore) Collection(ctx context.Context, db string, coll string) (*mongo.Collection, error) {
	collection := ms.client.Database(db).Collection(coll)
	fmt.Println(collection)
	return collection, nil
}

func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	return client, nil
}
