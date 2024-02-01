package store

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type Mongo interface {
	Collection(ctx context.Context, db, collection string) *mongo.Collection
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
