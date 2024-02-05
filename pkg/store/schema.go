package store

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Item struct {
	Id          primitive.ObjectID `bson:"_id"`
	Images      []string           `bson:"images"`
	Name        string             `bson:"name" validate:"required"`
	Price       float64            `bson:"price" validate:"required"`
	Summary     string             `bson:"summary" validate:"required"`
	Category    string             `bson:"category" validate:"required,oneof=beverages snacks"`
	Thumbnail   string             `bson:"thumbnail"`
	Description string             `bson:"description" validate:"required"`
	Ingridients []string           `bson:"ingridients" validate:"required"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

type ItemList []Item
