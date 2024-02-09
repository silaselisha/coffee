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
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/token"
	"github.com/silaselisha/coffee-api/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	hashedPassword := util.PasswordEncryption([]byte(data.Password))
	data.Id = primitive.NewObjectID()
	data.Role = "user"
	data.Avatar = "default.jpeg"
	data.Password = hashedPassword
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	_, err = collection.InsertOne(ctx, data)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	jwtoken := token.NewToken(s.envs.SecretAccessKey)
	minutes, err := strconv.Atoi(s.envs.JwtExpiresAt)
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
	}

	duration := time.Minute * time.Duration(minutes)
	token, err := jwtoken.CreateToken(ctx, data.Email, duration)
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
	}

	result := struct {
		Status string
		Token  string
		Data   store.User
	}{
		Status: "success",
		Token:  token,
		Data:   data,
	}
	return util.ResponseHandler(w, result, http.StatusCreated)
}
