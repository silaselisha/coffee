package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/internal"
	middleware "github.com/silaselisha/coffee-api/pkg/server/internal"
)

func productRoutes(gmux *mux.Router, srv *Server) {
	getItemsRouter := gmux.Methods(http.MethodGet).Subrouter()
	postItemsRouter := gmux.Methods(http.MethodPost).Subrouter()
	deleteItemsRouter := gmux.Methods(http.MethodDelete).Subrouter()
	updateItemsRouter := gmux.Methods(http.MethodPut).Subrouter()

	postItemsRouter.Use(middleware.AuthMiddleware(srv.Token))
	postProductsRouter := postItemsRouter.PathPrefix("/").Subrouter()
	postProductsRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin"))
	postProductsRouter.HandleFunc("/products", internal.HandleFuncDecorator(srv.CreateProductHandler))

	getItemsRouter.HandleFunc("/products", internal.HandleFuncDecorator(srv.GetAllProductsHandler))
	getItemsRouter.HandleFunc("/products/{category}/{id}", internal.HandleFuncDecorator(srv.GetProductByIdHandler))

	deleteItemsRouter.Use(middleware.AuthMiddleware(srv.Token))
	deleteProductsRouter := deleteItemsRouter.PathPrefix("/products").Subrouter()
	deleteProductsRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin"))
	deleteProductsRouter.HandleFunc("/{id}", internal.HandleFuncDecorator(srv.DeleteProductByIdHandler))

	updateItemsRouter.Use(middleware.AuthMiddleware(srv.Token))
	updateProductsRouter := updateItemsRouter.PathPrefix("/products").Subrouter()
	updateProductsRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin"))
	updateProductsRouter.HandleFunc("/{id}", internal.HandleFuncDecorator(srv.UpdateProductHandler))
}


func userRoutes(gmux *mux.Router, srv *Server) {
	userGetRouter := gmux.Methods(http.MethodGet).Subrouter()
	postUserRouter := gmux.Methods(http.MethodPost).Subrouter()
	forgotPasswordRouter := gmux.Methods(http.MethodPost).Subrouter()
	updateUserRouter := gmux.Methods(http.MethodPut).Subrouter()
	resetPasswordRouter := gmux.Methods(http.MethodPut).Subrouter()
	deleteUserRouter := gmux.Methods(http.MethodDelete).Subrouter()

	userGetRouter.Use(middleware.AuthMiddleware(srv.Token))

	getAllUsersRouter := userGetRouter.PathPrefix("/").Subrouter()
	getAllUsersRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin"))
	getAllUsersRouter.HandleFunc("/users", internal.HandleFuncDecorator(srv.GetAllUsersHandlers))

	getUserByIdRouter := userGetRouter.PathPrefix("/").Subrouter()
	getUserByIdRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin", "user"))
	getUserByIdRouter.HandleFunc("/users/{id}", internal.HandleFuncDecorator(srv.GetUserByIdHandler))

	postUserRouter.HandleFunc("/signup", internal.HandleFuncDecorator(srv.CreateUserHandler))
	postUserRouter.HandleFunc("/login", internal.HandleFuncDecorator(srv.LoginUserHandler))

	updateUserRouter.Use(middleware.AuthMiddleware(srv.Token))
	updateUserRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin", "user"))
	updateUserRouter.HandleFunc("/users/{id}", internal.HandleFuncDecorator(srv.UpdateUserByIdHandler))

	deleteUserRouter.Use(middleware.AuthMiddleware(srv.Token))
	deleteUserRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin", "user"))
	deleteUserRouter.HandleFunc("/users/{id}", internal.HandleFuncDecorator(srv.DeleteUserByIdHandler))
	forgotPasswordRouter.HandleFunc("/forgotpassword", internal.HandleFuncDecorator(srv.ForgotPasswordHandler))
	resetPasswordRouter.HandleFunc("/resetpassword", internal.HandleFuncDecorator(srv.ResetPasswordHandler))
}

func orderRoutes(gmux *mux.Router, srv *Server) {
	orderRouter := gmux.Methods(http.MethodPost).Subrouter()
	orderRouter.Use(middleware.AuthMiddleware(srv.Token))
	orderRouter.Use(middleware.RestrictToMiddleware(srv.Store, "user", "admin"))
	orderRouter.HandleFunc("/products/orders", internal.HandleFuncDecorator(srv.CreateOrderHandler))
}
