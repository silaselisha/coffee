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
	Author      primitive.ObjectID `bson:"author"`
	Thumbnail   string             `bson:"thumbnail"`
	Description string             `bson:"description" validate:"required"`
	Ingridients []string           `bson:"ingridients" validate:"required"`
	Ratings     float64            `bson:"ratings"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

type User struct {
	Id                primitive.ObjectID `bson:"_id"`
	Avatar            string             `bson:"avatar"`
	UserName          string             `bson:"username" validate:"required"`
	Role              string             `bson:"role"`
	Email             string             `bson:"email" validate:"required"`
	PhoneNumber       string             `bson:"phoneNumber" validate:"required"`
	Verified          bool               `bson:"verified"`
	Password          string             `bson:"password" validate:"required"`
	PasswordChangedAt time.Time          `bson:"password_changed_at,omitempty"`
	CreatedAt         time.Time          `bson:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at"`
}

type Reservation struct {
	Id        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type OrderStatus int32

const (
	PENDING OrderStatus = iota
	REJECTED
	PAID
)

type ItemOrder struct {
	Item        primitive.ObjectID `bson:"item"`
	Quantity    int64              `bson:"quantity"`
	Discount    float64            `bson:"discount"`
}

type Order struct {
	Id            primitive.ObjectID `bson:"_id"`
	User          primitive.ObjectID `bson:"user"`
	Items         []ItemOrder        `bosn:"items"`
	TotalDiscount float64            `bson:"total_discount"`
	TotalAmount   float64            `bson:"total_amount"`
	Status        OrderStatus        `bson:"status"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
}

type CoffeeDateTable struct{}
type Invoice struct{}

type ItemList []Item
type UserList []User
