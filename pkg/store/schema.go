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

type ItemUpdateParams struct {
	Images      []string  `bson:"images"`
	Name        string    `bson:"name"`
	Price       float64   `bson:"price"`
	Summary     string    `bson:"summary"`
	Category    string    `bson:"category" validate:"oneof=beverages snacks"`
	Thumbnail   string    `bson:"thumbnail"`
	Description string    `bson:"description"`
	Ingridients []string  `bson:"ingridients"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

type User struct {
	Id          primitive.ObjectID `bson:"_id"`
	Avatar      string             `bson:"avatar"`
	UserName    string             `bson:"username"`
	Email       string             `bson:"email"`
	PhoneNumber int                `bson:"phoneNumber"`
	Verified    bool               `bson:"verified"`
	Password    string             `bson:"password"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

type Reservation struct{}
type Order struct{}
type CoffeeDateTable struct{}
type Invoice struct{}

type ItemList []Item
type UserList []User
