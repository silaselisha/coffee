package products

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

type UpdateProductParams struct {
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Images      []string `json:"iamges"`
	Summary     string   `json:"summary"`
	Category    string   `json:"category"`
	Thumbnail   string   `json:"thumbnail"`
	Description string   `json:"Description"`
	Ingridients []string `json:"ingridients"`
}

func NewProduct(storage store.Store) ProductStore {
	return &ProductHandler{
		db: storage,
	}
}

func (h *ProductHandler) UpdateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection, err := h.db.Collection(ctx, "coffeeshop", "products", "")
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}
	dataBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	var data UpdateProductParams
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	// update data dynamically
	var updatedDocument Item
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$set", Value: data}}

	err = collection.FindOneAndUpdate(ctx, filter, update).Decode(&updatedDocument)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, err, http.StatusBadRequest)
		}
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	return nil
}

func (h *ProductHandler) GetAllProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection, err := h.db.Collection(ctx, "coffeeshop", "products", "")
	if err != nil {
		return util.ResponseHandler(w, "internal server error", http.StatusInternalServerError)
	}

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}
	defer cur.Close(ctx)

	var result ItemList
	for cur.Next(ctx) {
		item := new(Item)
		err := cur.Decode(&item)
		if err != nil {
			return util.ResponseHandler(w, err, http.StatusInternalServerError)
		}

		result = append(result, *item)
	}

	return util.ResponseHandler(w, result, http.StatusOK)
}

func (h *ProductHandler) GetProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection, err := h.db.Collection(ctx, "coffeeshop", "products", "")
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	filter := bson.D{{Key: "_id", Value: id}, {Key: "category", Value: vars["category"]}}
	cur := collection.FindOne(ctx, filter)

	var result Item
	err = cur.Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "document not found", http.StatusNotFound)
		}
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	return util.ResponseHandler(w, result, http.StatusOK)
}

func (h *ProductHandler) DeleteProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection, err := h.db.Collection(ctx, "coffeeshop", "products", "")
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}
	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	filter := bson.D{{Key: "_id", Value: id}}
	var deletedDocument bson.M
	err = collection.FindOneAndDelete(ctx, filter).Decode(&deletedDocument)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "invalid operation on data", http.StatusBadRequest)
		}
		return util.ResponseHandler(w, "internal server error", http.StatusInternalServerError)
	}

	return util.ResponseHandler(w, "", http.StatusNoContent)
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
