package store

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

type Store interface {
	Collection(ctx context.Context, db string, collection string) (*mongo.Collection, error)
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
