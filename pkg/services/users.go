// User services
// implementation SSO, socials signup & signin and traditional
// user signup & login
// cookie session management and refresh tokens
// password reset functionality; forgot apssword?
// account update info functionality [Avatar, PhoneNumber, Password]
// deactivate account and set for permanet deletion 30day without login
// delete account with all it's associated data
package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Server) CreateUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.db.Collection(ctx, "coffeeshop", "users")
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}, {Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	var data store.User
	userBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}
	err = json.Unmarshal(userBytes, &data)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	hashedPassword := fmt.Sprintf("%x", sha256.Sum256([]byte(data.Password)))
	data.Role = "user"
	data.Avatar = "default.jpeg"
	data.Password = hashedPassword
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	_, err = collection.InsertOne(ctx, data)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}
	return util.ResponseHandler(w, "user", http.StatusCreated)
}
