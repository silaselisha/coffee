package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UpdateProductParams struct {
	Name        string   `bson:"name"`
	Price       float64  `bson:"price"`
	Images      []string `bson:"iamges"`
	Summary     string   `bson:"summary"`
	Category    string   `bson:"category"`
	Thumbnail   string   `bson:"thumbnail"`
	Description string   `bson:"Description"`
	Ingridients []string `bson:"ingridients"`
}

func (h *Server) UpdateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := h.db.Collection(ctx, "coffeeshop", "products")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		return util.ResponseHandler(w, "invalid product", http.StatusBadRequest)
	}

	dataBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	var updatedDocument store.Item
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.M{"$set": data}

	err = collection.FindOneAndUpdate(ctx, filter, update).Decode(&updatedDocument)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, err, http.StatusNotFound)
		}
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	return nil
}

func (h *Server) GetAllProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := h.db.Collection(ctx, "coffeeshop", "products")

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}
	defer cur.Close(ctx)

	var result store.ItemList
	for cur.Next(ctx) {
		item := new(store.Item)
		err := cur.Decode(&item)
		if err != nil {
			return util.ResponseHandler(w, err, http.StatusInternalServerError)
		}

		result = append(result, *item)
	}

	return util.ResponseHandler(w, result, http.StatusOK)
}

func (h *Server) GetProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := h.db.Collection(ctx, "coffeeshop", "products")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	filter := bson.D{{Key: "_id", Value: id}, {Key: "category", Value: vars["category"]}}
	cur := collection.FindOne(ctx, filter)

	var result store.Item
	err = cur.Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "document not found", http.StatusNotFound)
		}
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	return util.ResponseHandler(w, result, http.StatusOK)
}

func (h *Server) DeleteProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := h.db.Collection(ctx, "coffeeshop", "products")

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

func (h *Server) CreateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := h.db.Collection(ctx, "coffeeshop", "products")

	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	fileName, err := util.ImageProcessor("images", r)
	if err != nil {
		log.Print(err)
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	return util.ResponseHandler(w, fileName, http.StatusOK)
}
