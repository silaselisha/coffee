package products

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductStore interface {
	CreateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	UpdateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	DeleteProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	GetAllProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	GetProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

type ProductHandler struct {
	db store.Store
}

func NewProduct(storage store.Store) ProductStore {
	return &ProductHandler{
		db: storage,
	}
}

func(h *ProductHandler) UpdateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func(h *ProductHandler) GetAllProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func(h *ProductHandler) GetProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}


func(h *ProductHandler) DeleteProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (h *ProductHandler) CreateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection, err := h.db.Collection(ctx, "coffeeshop", "products", "name")
	if err != nil {
		return util.ResponseHandler(w, "internal server error", http.StatusInternalServerError)
	}

	var data Item
	coffeeBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return util.ResponseHandler(w, "invalid read operation on data", http.StatusBadRequest)
	}

	err = json.Unmarshal(coffeeBytes, &data)
	if err != nil {
		return util.ResponseHandler(w, "invalid unmarshal operation on data", http.StatusBadRequest)
	}

	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	data.Id = primitive.NewObjectID()

	result, err := collection.InsertOne(ctx, data)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	return util.ResponseHandler(w, result, http.StatusCreated)
}
