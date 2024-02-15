package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Token interface {
	CreateToken(ctx context.Context, email string, duration time.Duration) (string, error)
	VerifyToken(ctx context.Context, token string) (*Payload, error)
}

type JWToken struct {
	secret string
}

func NewToken(secret string) Token {
	return &JWToken{
		secret: secret,
	}
}

func (tkn *JWToken) CreateToken(ctx context.Context, email string, duration time.Duration) (string, error) {
	payload, err := createNewPayload(duration, email)
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	tokenString, err := token.SignedString([]byte(tkn.secret))
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return tokenString, nil
}

func (tkn *JWToken) VerifyToken(ctx context.Context, tok string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("invalid jwt token signing alg")
		}
		return []byte(tkn.secret), nil
	}

	token, err := jwt.ParseWithClaims(tok, &Payload{}, keyFunc)
	if err != nil {
		varErr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(varErr, fmt.Errorf("invalid token")) {
			return nil, fmt.Errorf("invalid token")
		}
		return nil, err
	}

	payload, ok := token.Claims.(*Payload)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}

	return payload, nil
}
