package services

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type coffee struct {
	Id          primitive.ObjectID `bson:"_id"`
	Name        string             `bson:"name"`
	Country     string             `bson:"country"`
	Weight      int32              `bson:"weight"`
	Manufacture string             `bson:"manufacture"`
	Grade       int32              `bson:"grade"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}
