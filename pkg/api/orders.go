package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/internal"
	"github.com/silaselisha/coffee-api/pkg/middleware"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Server) CreateOrderHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ordColl := s.Store.Collection(ctx, "coffeeshop", "orders")
	prodColl := s.Store.Collection(ctx, "coffeeshop", "products")

	orderBytes, err := io.ReadAll(r.Body)
	if err != nil {
		response := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, response, http.StatusInternalServerError)
	}

	var orderReq types.OrderParams
	if err := json.Unmarshal(orderBytes, &orderReq); err != nil {
		response := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, response, http.StatusBadRequest)
	}

	userInfo := ctx.Value(types.AuthUserInfoKey{}).(*types.UserInfo)

	var products []store.OrderItem
	var totalAmount float64
	var totalDiscount float64

	for _, order := range orderReq.Items {
		id, err := primitive.ObjectIDFromHex(order.Product)
		if err != nil {
			response := internal.NewErrorResponse("failed", err.Error())
			return internal.ResponseHandler(w, response, http.StatusBadRequest)
		}

		var item store.Item
		err = prodColl.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&item)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				response := internal.NewErrorResponse("failed", err.Error())
				return internal.ResponseHandler(w, response, http.StatusBadRequest)
			}
			response := internal.NewErrorResponse("failed", err.Error())
			return internal.ResponseHandler(w, response, http.StatusInternalServerError)
		}

		amount := item.Price * float64(order.Quantity)
		discount := amount * float64(item.Discount) / 100.00

		product := store.OrderItem{
			Product:  item.Id,
			Quantity: order.Quantity,
			Amount:   amount,
			Discount: discount,
		}

		products = append(products, product)
		totalDiscount += discount

		totalAmount += (amount - totalDiscount)
	}

	orderPayload := store.Order{
		Id:            primitive.NewObjectID(),
		Items:         products,
		Owner:         userInfo.Id,
		Status:        "pending",
		TotalAmount:   totalAmount,
		TotalDiscount: totalDiscount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_, err = ordColl.InsertOne(ctx, orderPayload)
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
	orderRouter.HandleFunc("/products/orders", internal.HandleFuncDecorator(srv.CreateOrderHandler))
}
