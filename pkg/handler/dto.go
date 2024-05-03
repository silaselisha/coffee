package handler

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userLoginParams struct {
	Email    string `bson:"email" validate:"required"`
	Password string `bson:"password" validate:"required"`
}

type itemResponseParams struct {
	Id          string             `json:"_id"`
	Images      []string           `json:"images"`
	Name        string             `json:"name"`
	Author      primitive.ObjectID `json:"author"`
	Price       float64            `json:"price"`
	Summary     string             `json:"summary"`
	Category    string             `json:"category"`
	Thumbnail   string             `json:"thumbnail"`
	Description string             `json:"description"`
	Ingridients []string           `json:"ingridients"`
	Ratings     float64            `json:"ratings"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type itemResponseListParams []itemResponseParams
type passwordResetParams struct {
	Password        string `bson:"password" validate:"required"`
	ConfirmPassword string `bson:"confirmPassword" validate:"required"`
}
type forgotPasswordParams struct {
	Email string `bson:"email" validate:"required"`
}

type errorResponseParams struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func newErrorResponse(status string, err string) *errorResponseParams {
	return &errorResponseParams{
		Status: status,
		Error:  err,
	}
}
