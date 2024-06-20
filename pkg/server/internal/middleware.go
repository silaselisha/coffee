package internal

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/token"
	"github.com/silaselisha/coffee-api/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AuthMiddleware(tkn token.Token) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorizationHeader := r.Header.Get("authorization")
			if len(authorizationHeader) == 0 {
				http.Error(w, "invalid token header", http.StatusForbidden)
				return
			}

			fields := strings.Split(authorizationHeader, " ")
			if len(fields) < 2 {
				http.Error(w, "invalid token header", http.StatusForbidden)
				return
			}

			if strings.ToLower(fields[0]) != "bearer" {
				http.Error(w, "invalid token header", http.StatusForbidden)
				return
			}

			payload, err := tkn.VerifyToken(context.Background(), fields[1])
			if err != nil {
				http.Error(w, "invalid token", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), types.AuthPayloadKey{}, payload)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func RestrictToMiddleware(str store.Mongo, args ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var isAuthorized bool
			var authorized map[string]string = map[string]string{}
			roles := map[string]string{
				"admin": "admin",
				"user":  "user",
			}

			for _, role := range args {
				_, ok := roles[role]
				if ok {
					isAuthorized = true
					authorized[role] = role
				}
			}

			if !isAuthorized {
				err := errors.New("user forbidden to perform an operation on this resource 1")
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			payload := r.Context().Value(types.AuthPayloadKey{}).(*token.Payload)
			var user store.User
			collection := str.Collection(r.Context(), "coffeeshop", "users")
			id, err := primitive.ObjectIDFromHex(payload.Id)
			if err != nil {
				err := errors.New("user forbidden to perform an operation on this resource")
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			curr := collection.FindOne(r.Context(), bson.D{{Key: "_id", Value: id}})

			err = curr.Decode(&user)
			if err != nil {
				err := errors.New("user forbidden to perform an operation on this resource")
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			role, ok := authorized[user.Role]
			if !ok {
				err := errors.New("user forbidden to perform an operation on this resource")
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			var userInfo *types.UserInfo = &types.UserInfo{
				Role:   role,
				Email:  user.Email,
				Avatar: user.Avatar,
				Id:     id,
			}

			ctx := context.WithValue(r.Context(), types.AuthUserInfoKey{}, userInfo)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
