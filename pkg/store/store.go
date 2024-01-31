package store

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store interface {
	Collection(ctx context.Context, db, collection, field string) (*mongo.Collection, error)
}

type MongoStore struct {
	client *mongo.Client
}

func NewStore(client *mongo.Client) Store {
	return &MongoStore{
		client: client,
	}
}

func (ms *MongoStore) Collection(ctx context.Context, db, coll, field string) (*mongo.Collection, error) {
	collection := ms.client.Database(db).Collection(coll)
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: field, Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	if err != nil {
		return nil, err
	}
	return collection, nil
}
