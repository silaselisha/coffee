package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Payload struct {
	Email     string
	Id        uuid.UUID
	IssuedAt  time.Time
	ExpiredAt time.Time
}

func createNewPayload(duration time.Duration, email string) (*Payload, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("error creating a uuid %w", err)
	}
	return &Payload{
		Email:     email,
		Id:        id,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}, nil
}

func (p *Payload) Valid() error {
	if time.Now().After(p.ExpiredAt) {
		err := errors.New("invalid token")
		return err
	}
	return nil
}
