package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"github.com/silaselisha/coffee-api/pkg/workers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Server) UpdateProductHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "products")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		return util.ResponseHandler(w, "invalid product", http.StatusBadRequest)
	}

	err = r.ParseMultipartForm(int64(32 << 20))
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	fields := []string{"name", "price", "ingridients", "description", "summary"}
	data := bson.M{}
	for _, field := range fields {
		value := r.FormValue(field)
		if value != "" {
			data[field] = r.FormValue(field)
		}
	}

	if data["price"] != "" {
		price, err := strconv.ParseFloat(r.FormValue("price"), 64)
		if err != nil {
			return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
		}
		data["price"] = price
	}

	if data["ingridients"] != "" {
		var result []string = strings.Split(r.FormValue("ingridients"), ",")
		data["ingridients"] = result
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

	product := itemResponseParams{
		Id:          updatedDocument.Id.Hex(),
		Images:      updatedDocument.Images,
		Name:        updatedDocument.Name,
		Price:       updatedDocument.Price,
		Summary:     updatedDocument.Summary,
		Category:    updatedDocument.Category,
		Thumbnail:   updatedDocument.Thumbnail,
		Description: updatedDocument.Description,
		Ingridients: updatedDocument.Ingridients,
		Ratings:     updatedDocument.Ratings,
		CreatedAt:   updatedDocument.CreatedAt,
		UpdatedAt:   updatedDocument.UpdatedAt,
	}

	result := struct {
		Status string             `json:"status"`
		Data   itemResponseParams `json:"data"`
	}{
		Status: "success",
		Data:   product,
	}
	return util.ResponseHandler(w, result, http.StatusOK)
}

func (s *Server) GetAllProductsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "products")

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}
	defer cur.Close(ctx)

	var result itemResponseListParams
	for cur.Next(ctx) {
		item := new(store.Item)
		err := cur.Decode(&item)
		if err != nil {
			return util.ResponseHandler(w, err, http.StatusInternalServerError)
		}

		product := itemResponseParams{
			Id:          item.Id.Hex(),
			Images:      item.Images,
			Name:        item.Name,
			Price:       item.Price,
			Summary:     item.Summary,
			Category:    item.Category,
			Thumbnail:   item.Thumbnail,
			Description: item.Description,
			Ingridients: item.Ingridients,
			Ratings:     item.Ratings,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
		result = append(result, product)
	}

	resp := struct {
		Status  string                 `json:"status"`
		Results int32                  `json:"results"`
		Data    itemResponseListParams `json:"data"`
	}{
		Status:  "success",
		Results: int32(len(result)),
		Data:    result,
	}
	return util.ResponseHandler(w, resp, http.StatusOK)
}

func (s *Server) GetProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "products")

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

	product := itemResponseParams{
		Id:          item.Id.Hex(),
		Images:      item.Images,
		Name:        item.Name,
		Price:       item.Price,
		Summary:     item.Summary,
		Category:    item.Category,
		Thumbnail:   item.Thumbnail,
		Description: item.Description,
		Ingridients: item.Ingridients,
		Ratings:     item.Ratings,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}

	res := struct {
		Status string             `json:"status"`
		Data   itemResponseParams `json:"data"`
	}{
		Status: "success",
		Data:   product,
	}
	return util.ResponseHandler(w, res, http.StatusOK)
}

func (s *Server) DeleteProductByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "products")

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
	session, err := s.Store.TxnStartSession(ctx)
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
	}

	defer session.EndSession(ctx)
	resposne, err := session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		collection := s.Store.Collection(ctx, "coffeeshop", "products")
		_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			if err := session.AbortTransaction(ctx); err != nil {
				return nil, err
			}
			return nil, err
		}

		err = r.ParseMultipartForm(int64(32 << 20))
		if err != nil {
			if err := session.AbortTransaction(ctx); err != nil {
				return nil, err
			}
			return nil, err
		}

		var item store.Item
		price, err := strconv.ParseFloat(r.FormValue("price"), 64)
		if err != nil {
			if err := session.AbortTransaction(ctx); err != nil {
				return nil, err
			}
			return nil, err
		}

		var ingridients []string = strings.Split(r.FormValue("ingridients"), ",")

		file, _, err := r.FormFile("thumbnail")
		if err != nil {
			if err := session.AbortTransaction(ctx); err != nil {
				return nil, err
			}
			return nil, err
		}
		defer file.Close()

		thumbnail, fileName, extension, err := util.ImageProcessor(ctx, file, &util.FileMetadata{ContetntType: "image"})
		if err != nil {
			if err := session.AbortTransaction(ctx); err != nil {
				return nil, err
			}
			return nil, err
		}

		objectKey := fmt.Sprintf("images/products/thumbnails/%s", fileName)
		item = store.Item{
			Id:          primitive.NewObjectID(),
			Name:        r.FormValue("name"),
			Price:       price,
			Ingridients: ingridients,
			Thumbnail:   objectKey,
			Summary:     r.FormValue("summary"),
			Category:    r.FormValue("category"),
			Description: r.FormValue("description"),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		_, err = collection.InsertOne(ctx, item)
		if err != nil {
			if err := session.AbortTransaction(ctx); err != nil {
				return nil, err
			}
			return nil, err
		}

		opts := []asynq.Option{
			asynq.MaxRetry(3),
			asynq.ProcessIn(2 * time.Second),
			asynq.Queue(workers.CriticalQueue),
		}

		err = s.distributor.SendS3ObjectUploadTask(ctx, &workers.PayloadUploadImage{
			FileName:  fileName,
			ObjectKey: objectKey,
			Extension: extension,
			Image:     thumbnail,
		}, opts...)
		if err != nil {
			if err := session.AbortTransaction(ctx); err != nil {
				return nil, err
			}
			return nil, err
		}

		product := itemResponseParams{
			Id:          item.Id.Hex(),
			Name:        item.Name,
			Price:       item.Price,
			Ingridients: item.Ingridients,
			Thumbnail:   item.Thumbnail,
			Summary:     item.Summary,
			Category:    item.Category,
			Description: item.Description,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}

		err = session.CommitTransaction(ctx)
		if err != nil {
			return nil, err
		}
		return product, nil
	}, &options.TransactionOptions{})

	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
	}

	product := resposne.(itemResponseParams)
	result := struct {
		Status string             `json:"status"`
		Data   itemResponseParams `json:"data"`
	}{
		Status: "success",
		Data:   product,
	}

	return util.ResponseHandler(w, result, http.StatusCreated)
}

func productRoutes(gmux *mux.Router, srv *Server) {
	getProductRouter := gmux.Methods(http.MethodGet).Subrouter()
	postProductRouter := gmux.Methods(http.MethodPost).Subrouter()
	deleteProductRouter := gmux.Methods(http.MethodDelete).Subrouter()
	updateProductRouter := gmux.Methods(http.MethodPut).Subrouter()

	postProductRouter.HandleFunc("/products", util.HandleFuncDecorator(srv.CreateProductHandler))
	getProductRouter.HandleFunc("/products", util.HandleFuncDecorator(srv.GetAllProductsHandler))
	getProductRouter.HandleFunc("/products/{category}/{id}", util.HandleFuncDecorator(srv.GetProductByIdHandler))
	deleteProductRouter.HandleFunc("/products/{id}", util.HandleFuncDecorator(srv.DeleteProductByIdHandler))
	updateProductRouter.HandleFunc("/products/{id}", util.HandleFuncDecorator(srv.UpdateProductHandler))
}
