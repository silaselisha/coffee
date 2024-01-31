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

type ProductInter interface {
	CreateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

type Product struct {
	db store.Store
}

func NewProduct(storage store.Store) ProductInter {
	return &Product{
		db: storage,
	}
}

func (s *Product) CreateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection, err := s.db.Collection(ctx, "coffeeshop", "products", "name")
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
