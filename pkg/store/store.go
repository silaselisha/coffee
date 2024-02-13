package store

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

type Mongo interface {
	Collection(ctx context.Context, db, collection string) *mongo.Collection
	Disconnect(ctx context.Context) error
}

type MongoClient struct {
	client *mongo.Client
}

func NewMongoClient(client *mongo.Client) Mongo {
	return &MongoClient{
		client: client,
	}
}

func (ms *MongoClient) Collection(ctx context.Context, db, coll string) *mongo.Collection {
	collection := ms.client.Database(db).Collection(coll)
	return collection
}

func (ms *MongoClient) Disconnect(ctx context.Context) error {
	err := ms.client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("database disconnection error %w", err)
	}
	return nil
}