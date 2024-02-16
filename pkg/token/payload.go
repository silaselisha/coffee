package token

import (
	"errors"
	"time"
)

type Payload struct {
	Email     string
	Id        string
	IssuedAt  time.Time
	ExpiredAt time.Time
}

func createNewPayload(duration time.Duration, id, email string) (*Payload, error) {
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
