package services

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
)

type Server struct {
	Router *mux.Router
	db     store.Mongo
	vd     *validator.Validate
}

func NewServer(store store.Mongo) store.Querier {
	server := &Server{db: store}
	router := mux.NewRouter()

	validate := validator.New(validator.WithRequiredStructEnabled())
	server.vd = validate

	getRouter := router.Methods(http.MethodGet).Subrouter()
	postRouter := router.Methods(http.MethodPost).Subrouter()
	deleteRouter := router.Methods(http.MethodDelete).Subrouter()
	updateRouter := router.Methods(http.MethodPut).Subrouter()

	postRouter.HandleFunc("/products", util.HandleFuncDecorator(server.CreateProductHandler))
	getRouter.HandleFunc("/products", util.HandleFuncDecorator(server.GetAllProductHandler))
	getRouter.HandleFunc("/products/{category}/{id}", util.HandleFuncDecorator(server.GetProductByIdHandler))
	deleteRouter.HandleFunc("/products/{id}", util.HandleFuncDecorator(server.DeleteProductByIdHandler))
	updateRouter.HandleFunc("/products/{id}", util.HandleFuncDecorator(server.UpdateProductHandler))

	server.Router = router
	return server
}
