package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/silaselisha/coffee-api/pkg/token"
)

type AuthKey struct{}
type RolesKey struct{}

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

			ctx := context.WithValue(r.Context(), AuthKey{}, payload)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func RestrictToMiddleware(args ...string) func(next http.Handler) http.Handler {
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
				err := errors.New("user forbidden to access this resource")
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), RolesKey{}, authorized)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
