package handler

import (
	"time"
)

type userLoginParams struct {
	Email    string `bson:"email" validate:"required"`
	Password string `bson:"password" validate:"required"`
}

type userResponseParams struct {
	Id          string    `json:"_id"`
	Avatar      string    `json:"avatar"`
	UserName    string    `json:"username"`
	Role        string    `json:"role"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone"`
	Verified    bool      `json:"Verified"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type userResponseListParams []userResponseParams

type passwordResetParams struct {
	Password        string `bson:"password" validate:"required"`
	ConfirmPassword string `bson:"confirmPassword" validate:"required"`
}

type forgotPasswordParams struct {
	Email string `bson:"email" validate:"required"`
}
