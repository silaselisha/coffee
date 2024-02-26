package store

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo interface {
	Disconnect(ctx context.Context) error
	TxnStartSession(ctx context.Context) (mongo.Session, error)
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

func (ms *MongoClient) Disconnect(ctx context.Context) error {
	err := ms.client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("database disconnection error %w", err)
	}
	return nil
}

func (ms *MongoClient) TxnStartSession(ctx context.Context) (mongo.Session, error) {
	session, err := ms.client.StartSession(&options.SessionOptions{})
	if err != nil {
		return nil, err
	}

	// defer session.EndSession(ctx)
	return session, nil
}
