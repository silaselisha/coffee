package store

import (
	"context"
	"net/http"
)

type Querier interface {
	CreateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	UpdateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	DeleteProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	GetAllProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	GetProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

// var data store.Item
	// productBytes, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	return util.ResponseHandler(w, err, http.StatusBadRequest)
	// }

	// err = json.Unmarshal(productBytes, &data)
	// if err != nil {
	// 	return util.ResponseHandler(w, "invalid unmarshal operation on data", http.StatusBadRequest)
	// }

	// err = h.vd.Struct(data)
	// if err != nil {
	// 	return util.ResponseHandler(w, err, http.StatusBadRequest)
	// }

	// data.CreatedAt = time.Now()
	// data.UpdatedAt = time.Now()
	// data.Id = primitive.NewObjectID()

	// result, err := collection.InsertOne(ctx, data)
	// if err != nil {
	// 	return util.ResponseHandler(w, err, http.StatusInternalServerError)
	// }

	// return util.ResponseHandler(w, result, http.StatusCreated)