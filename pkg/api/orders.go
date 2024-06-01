package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/internal"
	"github.com/silaselisha/coffee-api/pkg/middleware"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/types"
)

func (s *Server) CreateOrderHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// TODO: Create a unique order and persit it to the the data Store

	fmt.Println("ORDER")
	collection := s.Store.Collection(ctx, "coffeeshop", "orders")
	orderBytes, err := io.ReadAll(r.Body)
	if err != nil {
		response := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, response, http.StatusInternalServerError)
	}

	var orderPayload store.Order
	if err := json.Unmarshal(orderBytes, &orderPayload); err != nil {
		response := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, response, http.StatusBadRequest)
	}

	userInfo := ctx.Value(types.AuthUserInfoKey{}).(*types.UserInfo)

	orderPayload.Owner = userInfo.Id
	for _, order := range orderPayload.Items {
		// TODO: coveret type string of the order ID into a primitive OBJECT ID

		fmt.Println(order)
	}
	_, err = collection.InsertOne(ctx, orderPayload)
	if err != nil {
		response := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, response, http.StatusInternalServerError)
	}
	return internal.ResponseHandler(w, orderPayload, http.StatusCreated)
}

func orderRoutes(gmux *mux.Router, srv *Server) {
	orderRouter := gmux.Methods(http.MethodPost).Subrouter()
	orderRouter.Use(middleware.AuthMiddleware(srv.token))
	orderRouter.Use(middleware.RestrictToMiddleware(srv.Store, "user", "admin"))
	orderRouter.HandleFunc("/orders", internal.HandleFuncDecorator(srv.CreateOrderHandler))
}
