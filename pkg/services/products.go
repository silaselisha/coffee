package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Server) UpdateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.db.Collection(ctx, "coffeeshop", "products")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		return util.ResponseHandler(w, "invalid product", http.StatusBadRequest)
	}

	err = r.ParseMultipartForm(int64(32 << 20))
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	ingridients := strings.Split(r.FormValue("ingridients"), ",")

	data := store.ItemUpdateParams{
		Name:        r.FormValue("name"),
		Price:       price,
		Description: r.FormValue("description"),
		Summary:     r.FormValue("summary"),
		Ingridients: ingridients,
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

	result := struct {
		Status string
		Data   store.Item
	}{
		Status: "success",
		Data:   updatedDocument,
	}
	return util.ResponseHandler(w, result, http.StatusOK)
}

func (s *Server) GetAllProductsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.db.Collection(ctx, "coffeeshop", "products")

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

	resp := struct {
		Status  string
		Results int32
		Data    store.ItemList
	}{
		Status:  "success",
		Results: int32(len(result)),
		Data:    result,
	}
	return util.ResponseHandler(w, resp, http.StatusOK)
}

func (s *Server) GetProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.db.Collection(ctx, "coffeeshop", "products")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	filter := bson.D{{Key: "_id", Value: id}, {Key: "category", Value: vars["category"]}}
	result := collection.FindOne(ctx, filter)

	var item store.Item
	err = result.Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "document not found", http.StatusNotFound)
		}
		return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
	}

	res := struct {
		Status string
		Data   store.Item
	}{
		Status: "success",
		Data:   item,
	}
	return util.ResponseHandler(w, res, http.StatusOK)
}

func (s *Server) DeleteProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.db.Collection(ctx, "coffeeshop", "products")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
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

func (s *Server) CreateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.db.Collection(ctx, "coffeeshop", "products")

	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	err = r.ParseMultipartForm(int64(32 << 20))
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	thumbnailName := make(chan string)
	errs := make(chan error)
	go func() {
		file, _, err := r.FormFile("thumbnail")
		if err != nil {
			errs <- err
			return
		}

		_, fileName, err := util.ImageThumbnailProcessor(ctx, file)
		if err != nil {
			errs <- err
			return
		}
		
		thumbnailName <- fileName
		close(thumbnailName)
		defer file.Close()
	}()

	var item store.Item
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	select {
	case err := <-errs:
		if err != nil {
			fmt.Println(err.Error())
			return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
		}

	case fileName, ok := <-thumbnailName:
		if !ok {
			break
		}
		item.Thumbnail = fileName
	}

	var ingridients []string = strings.Split(r.FormValue("ingridients"), ",")

	item = store.Item{
		Id:          primitive.NewObjectID(),
		Name:        r.FormValue("name"),
		Price:       price,
		Ingridients: ingridients,
		Summary:     r.FormValue("summary"),
		Category:    r.FormValue("category"),
		Description: r.FormValue("description"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = collection.InsertOne(ctx, item)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	result := struct {
		Status string
		Data   store.Item
	}{
		Status: "success",
		Data:   item,
	}

	return util.ResponseHandler(w, result, http.StatusCreated)
}
