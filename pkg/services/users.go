package services

import (
	"context"
	"net/http"

	"github.com/silaselisha/coffee-api/pkg/util"
)

func(h *Server) CreateUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return util.ResponseHandler(w, "user", http.StatusCreated)
}