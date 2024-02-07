package services

import (
	"context"
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

type itemParams struct {
	id          primitive.ObjectID `bson:"_id"`
	name        string             `bson:"name"`
	price       float64            `bson:"price"`
	description string             `bson:"description"`
	summary     string             `bson:"summary"`
	thumbnail   string             `bson:"thumbnail"`
	category    string             `bson:"category"`
	images      []string           `bson:"images"`
	ingridients []string           `bson:"ingridients"`
	created_at  time.Time          `bson:"created_at"`
	updated_at  time.Time          `bson:"updated_at"`
}

func (h *Server) UpdateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := h.db.Collection(ctx, "coffeeshop", "products")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		return util.ResponseHandler(w, "invalid product", http.StatusBadRequest)
	}

	err = r.ParseMultipartForm(int64(32 << 20))
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	thumbnail := make(chan string, 1)
	errs := make(chan error, 1)
	if file, _, err := r.FormFile("thumbnail"); err == nil {
		go func() {
			filename, err := util.S3ImageUploader(ctx, file)
			if err != nil {
				errs <- err
				return
			}
			thumbnail <- filename
		}()
	}

	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	ingridients := strings.Split(r.FormValue("ingridients"), ",")

	data := store.ItemUpdateParams{
		Name:        r.FormValue("name"),
		Price:       price,
		Thumbnail:   <-thumbnail,
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
		status string
		data   itemParams
	}{
		status: "success",
		data: itemParams{
			id:          updatedDocument.Id,
			name:        updatedDocument.Name,
			price:       updatedDocument.Price,
			summary:     updatedDocument.Summary,
			category:    updatedDocument.Category,
			images:      updatedDocument.Images,
			thumbnail:   updatedDocument.Thumbnail,
			ingridients: updatedDocument.Ingridients,
			description: updatedDocument.Description,
			created_at:  updatedDocument.CreatedAt,
			updated_at:  updatedDocument.UpdatedAt,
		},
	}
	return util.ResponseHandler(w, result, http.StatusOK)
}

func (h *Server) GetAllProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := h.db.Collection(ctx, "coffeeshop", "products")

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}
	defer cur.Close(ctx)

	var result []itemParams
	for cur.Next(ctx) {
		item := new(itemParams)
		err := cur.Decode(&item)
		if err != nil {
			return util.ResponseHandler(w, err, http.StatusInternalServerError)
		}

		result = append(result, *item)
	}

	resp := struct {
		status  string
		results int32
		data    []itemParams
	}{
		status:  "success",
		results: int32(len(result)),
		data:    result,
	}
	return util.ResponseHandler(w, resp, http.StatusOK)
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

	var item itemParams
	err = cur.Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "document not found", http.StatusNotFound)
		}
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	result := struct {
		status string
		data   itemParams
	}{
		status: "success",
		data:   item,
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

	err = r.ParseMultipartForm(int64(32 << 20))
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	thumbnail := make(chan string, 1)
	errs := make(chan error, 1)
	go func() {
		file, _, err := r.FormFile("thumbnail")
		if err != nil {
			errs <- err
			return
		}
		data, err := util.S3ImageUploader(ctx, file)
		if err != nil {
			errs <- err
			return
		}
		thumbnail <- data
		close(thumbnail)
	}()

	var item store.Item
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	select {
	case filename := <-thumbnail:
		item.Thumbnail = filename
	case err := <-errs:
		if err != nil {
			return util.ResponseHandler(w, err, http.StatusBadRequest)
		}
	}

	var ingridients []string = strings.Split(r.FormValue("ingridients"), ",")

	item = store.Item{
		Name:        r.FormValue("name"),
		Price:       price,
		Ingridients: ingridients,
		Summary:     r.FormValue("summary"),
		Category:    r.FormValue("category"),
		Description: r.FormValue("description"),
		Images:      nil,
		Thumbnail:   "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	resultId, err := collection.InsertOne(ctx, item)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	result := struct {
		status string
		data   itemParams
	}{
		status: "success",
		data: itemParams{
			id:          resultId.InsertedID.(primitive.ObjectID),
			name:        item.Name,
			price:       item.Price,
			summary:     item.Summary,
			category:    item.Category,
			images:      item.Images,
			thumbnail:   item.Thumbnail,
			ingridients: item.Ingridients,
			description: item.Description,
			created_at:  item.CreatedAt,
			updated_at:  item.UpdatedAt,
		},
	}

	return util.ResponseHandler(w, result, http.StatusCreated)
}
