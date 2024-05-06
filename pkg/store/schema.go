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
	Discount    uint32             `bson:"discount"`
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

type OrderItem struct {
	Product  primitive.ObjectID `bson:"product"`
	Quantity uint32             `bson:"quantity"`
	Amount   float64            `bson:"amount"`
}

type Order struct {
	Id            primitive.ObjectID `bson:"_id"`
	Items         []OrderItem        `bson:"items"`
	TotalAmount   float64            `bson:"total_amount"`
	Owner         primitive.ObjectID `bson:"owner"`
	Status        string             `bson:"status"`
	TotalDiscount float64            `bson:"total_discount"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
}

type CoffeeDateTable struct{}
type Invoice struct{}

type ItemList []Item
type UserList []User
