package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/silaselisha/coffee-api/internal"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Server) CreateOrderHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ordColl := s.Store.Collection(ctx, "coffeeshop", "orders")

	orderPayload, err := internal.ReadReqBody[types.OrderParams](r.Body, s.vd)
	if err != nil {
		res := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, res, http.StatusBadRequest)
	}

	userInfo := ctx.Value(types.AuthUserInfoKey{}).(*types.UserInfo)

	productsID, cart, err := internal.ExtractProductsID(orderPayload)
	if err != nil {
		res := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, res, http.StatusInternalServerError)
	}

	products, err := s.BatchGetAllProductsByIds(ctx, productsID)
	if err != nil {
		res := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, res, http.StatusInternalServerError)
	}

	var totalAmount float64
	var totalDiscount float64
	var orderItems []store.OrderItem
	for _, order := range cart {
		product, ok := products[order.Product]

		if !ok {
			res := internal.NewErrorResponse("failed", fmt.Errorf("product not found").Error())
			return internal.ResponseHandler(w, res, http.StatusNotFound)
		}

		amount := product.Price * float64(order.Quantity)
		discount := (amount / 100.00) * float64(product.Discount)
		totalAmount += (amount - discount)
		totalDiscount += discount

		orderItem := store.OrderItem{
			Product:  order.Product,
			Quantity: order.Quantity,
			Amount:   amount,
			Discount: discount,
		}
		orderItems = append(orderItems, orderItem)
	}

	order := store.Order{
		Id:            primitive.NewObjectID(),
		Items:         orderItems,
		TotalAmount:   totalAmount,
		Owner:         userInfo.Id,
		Status:        "pending",
		TotalDiscount: totalDiscount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = ordColl.InsertOne(ctx, order)
	if err != nil {
		res := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, res, http.StatusInternalServerError)
	}

	return internal.ResponseHandler(w, order, http.StatusCreated)
}
