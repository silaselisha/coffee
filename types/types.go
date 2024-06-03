package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentStatus int

const (
	PAID PaymentStatus = iota
	PENDING
	REJECTED
)

type FileMetadata struct {
	ContetntType string
}

type PayloadUploadImage struct {
	Image     []byte `json:"image"`
	ObjectKey string `json:"objectKey"`
	Extension string `json:"extension"`
}

type PayloadSendMail struct {
	Email string `json:"email"`
}

type UserReqParams struct {
	UserName    string `bson:"username" validate:"required"`
	Email       string `bson:"email" validate:"required"`
	PhoneNumber string `bson:"phoneNumber" validate:"required"`
	Password    string `bson:"password" validate:"required"`
}

type UserResParams struct {
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

type UserResListParams []UserResParams

type ItemResParams struct {
	Id          string             `json:"_id"`
	Images      []string           `json:"images"`
	Name        string             `json:"name"`
	Author      primitive.ObjectID `json:"author"`
	Price       float64            `json:"price"`
	Discount    uint32             `json:"discount"`
	Summary     string             `json:"summary"`
	Category    string             `json:"category"`
	Thumbnail   string             `json:"thumbnail"`
	Description string             `json:"description"`
	Ingridients []string           `json:"ingridients"`
	Ratings     float64            `json:"ratings"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type ItemResponseListParams []ItemResParams

type AuthPayloadKey struct{}
type AuthUserInfoKey struct{}
type UserInfo struct {
	Role   string
	Email  string
	Avatar string
	Id     primitive.ObjectID
}

type UserLoginParams struct {
	Email    string `bson:"email" validate:"required"`
	Password string `bson:"password" validate:"required"`
}

type PasswordResetParams struct {
	Password        string `bson:"password" validate:"required"`
	ConfirmPassword string `bson:"confirmPassword" validate:"required"`
}
type ForgotPasswordParams struct {
	Email string `bson:"email" validate:"required"`
}

type ErrorResParams struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type OrderItemParams struct {
	Product  string  `bson:"product"`
	Quantity uint32  `bson:"quantity"`
	Amount   float64 `bson:"amount"`
	Discount float64 `bson:"discount"`
}

type OrderParams struct {
	Items []OrderItemParams `bson:"items" validate:"required"`
}

type Config struct {
	DB_URI               string `mapstructure:"DB_URI"`
	SMTP_HOST            string `mapstructure:"SMTP_HOST"`
	SMTP_PORT            string `mapstructure:"SMTP_PORT"`
	DB_PASSWORD          string `mapstructure:"DB_PASSWORD"`
	SMTP_PASSWORD        string `mapstructure:"SMTP_PASSWORD"`
	SMTP_USERNAME        string `mapstructure:"SMTP_USERNAME"`
	S3_BUCKET_NAME       string `mapstructure:"S3_BUCKET_NAME"`
	SMTP_SENDER          string `mapstructure:"SMTP_SENDER"`
	SERVER_REST_ADDRESS  string `mapstructure:"SERVER_REST_ADDRESS"`
	JWT_EXPIRES_AT       string `mapstructure:"JWT_EXPIRES_AT"`
	SECRET_ACCESS_KEY    string `mapstructure:"SECRET_ACCESS_KEY"`
	REDIS_SERVER_PORT    string `mapstructure:"REDIS_SERVER_PORT"`
	REDIS_SERVER_ADDRESS string `mapstructure:"REDIS_SERVER_ADDRESS"`
}
