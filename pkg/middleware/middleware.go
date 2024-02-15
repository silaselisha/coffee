package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/silaselisha/coffee-api/pkg/token"
)

var AuthUser struct{} = struct{}{}

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

			_, err := tkn.VerifyToken(context.Background(), fields[1])
			if err != nil {
				http.Error(w, "invalid token", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
