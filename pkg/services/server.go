package services

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	Router *mux.Router
	db     store.Mongo
	vd     *validator.Validate
	envs   *util.Config
}

func NewServer(client *mongo.Client) store.Querier {
	server := &Server{}

	envs, err := util.LoadEnvs("./../../")
	if err != nil {
		log.Panic(err)
	}

	store := store.NewMongoClient(client)
	server.db = store
	server.envs = envs

	validate := validator.New(validator.WithRequiredStructEnabled())
	server.vd = validate

	router := mux.NewRouter()
	getProductRouter := router.Methods(http.MethodGet).Subrouter()
	postProductRouter := router.Methods(http.MethodPost).Subrouter()
	deleteProductRouter := router.Methods(http.MethodDelete).Subrouter()
	updateProductRouter := router.Methods(http.MethodPut).Subrouter()

	postUserRouter := router.Methods(http.MethodPost).Subrouter()

	postProductRouter.HandleFunc("/products", util.HandleFuncDecorator(server.CreateProductHandler))
	getProductRouter.HandleFunc("/products", util.HandleFuncDecorator(server.GetAllProductHandler))
	getProductRouter.HandleFunc("/products/{category}/{id}", util.HandleFuncDecorator(server.GetProductByIdHandler))
	deleteProductRouter.HandleFunc("/products/{id}", util.HandleFuncDecorator(server.DeleteProductByIdHandler))
	updateProductRouter.HandleFunc("/products/{id}", util.HandleFuncDecorator(server.UpdateProductHandler))

	postUserRouter.HandleFunc("/users/signup", util.HandleFuncDecorator(server.CreateUserHandler))

	server.Router = router
	return server
}
