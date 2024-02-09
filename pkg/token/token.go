package token

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Token interface {
	CreateToken(ctx context.Context, email string, duration time.Duration) (string, error)
	VerifyToken(ctx context.Context, token string) (bool, error)
}

type JWToken struct {
	secret string
}

func NewToken(secret string) Token {
	return &JWToken{
		secret: secret,
	}
}

func(j *JWToken) CreateToken(ctx context.Context, email string, duration time.Duration) (string, error) {
	payload, err := createNewPayload(duration, email)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	
	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return tokenString, nil
}

func(j *JWToken) VerifyToken(ctx context.Context, token string) (bool, error) {
	return true, nil
}
