// delete account and it's s3 object
// updating user profile image then delete the exisitng one
// resize the image (research)
// email user code URL for verification, reset password
// create a path to verify phone number
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/pkg/middleware"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/token"
	"github.com/silaselisha/coffee-api/pkg/util"
	"github.com/silaselisha/coffee-api/pkg/workers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Server) LoginUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	credentialsBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	var credentials userLoginParams
	json.Unmarshal(credentialsBytes, &credentials)
	err = s.vd.Struct(credentials)
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	var user store.User
	collection := s.Store.Collection(ctx, "coffeeshop", "users")
	curr := collection.FindOne(ctx, bson.D{{Key: "email", Value: credentials.Email}})
	if err := curr.Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "document not found", http.StatusNotFound)
		}
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	if !util.ComparePasswordEncryption(credentials.Password, user.Password) {
		return util.ResponseHandler(w, "invalid email or password", http.StatusBadRequest)
	}

	jwtToken := token.NewToken(s.envs.SECRET_ACCESS_KEY)
	days, err := strconv.Atoi(s.envs.JWT_EXPIRES_AT)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	hrs := fmt.Sprintf("%dh", (days * 24))
	duration, err := time.ParseDuration(hrs)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	token, err := jwtToken.CreateToken(ctx, duration, user.Id.Hex(), user.Email)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}
	res := struct {
		Status string
		Token  string
	}{
		Status: "success",
		Token:  token,
	}
	return util.ResponseHandler(w, res, http.StatusOK)
}

func (s *Server) CreateUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "users")
	_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
	})

	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	var data store.User

	userBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}
	err = json.Unmarshal(userBytes, &data)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	if err := s.vd.Struct(data); err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	hashedPassword := util.PasswordEncryption([]byte(data.Password))
	data.Id = primitive.NewObjectID()
	data.Role = "user"
	data.Avatar = "default.jpeg"
	data.Password = hashedPassword
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	_, err = collection.InsertOne(ctx, data)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(1 * time.Minute),
		asynq.Queue(workers.CriticalQueue),
	}
	err = s.distributor.SendMailTask(ctx, &workers.PayloadSendMail{Email: data.Email}, opts...)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	jwtoken := token.NewToken(s.envs.SECRET_ACCESS_KEY)
	days, err := strconv.Atoi(s.envs.JWT_EXPIRES_AT)
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
	}

	hrs := fmt.Sprintf("%dh", (days * 24))
	duration, err := time.ParseDuration(hrs)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	token, err := jwtoken.CreateToken(ctx, duration, data.Id.Hex(), data.Email)
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
	}

	result := struct {
		Status string
		Token  string
		Data   store.User
	}{
		Status: "success",
		Token:  token,
		Data:   data,
	}
	return util.ResponseHandler(w, result, http.StatusCreated)
}

func (s *Server) GetAllUsersHandlers(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "users")

	var users store.UserList
	curr, err := collection.Find(ctx, bson.D{{}})
	if err != nil {
		log.Print(err)
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	defer curr.Close(ctx)
	for curr.Next(ctx) {
		var user store.User
		err := curr.Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				break
			}

			return util.ResponseHandler(w, err, http.StatusInternalServerError)
		}
		users = append(users, user)
	}

	result := struct {
		Status string
		Result int32
		Data   store.UserList
	}{
		Status: "success",
		Result: int32(len(users)),
		Data:   users,
	}
	return util.ResponseHandler(w, result, http.StatusOK)
}

func (s *Server) GetUserByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "users")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	payload := ctx.Value(middleware.AuthPayloadKey{}).(*token.Payload)
	role := ctx.Value(middleware.AuthRoleKey{}).(string)

	if payload.Id != id.Hex() && role != "admin" {
		return util.ResponseHandler(w, "login or signup to perform this request", http.StatusForbidden)
	}

	curr := collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}})
	var user store.User
	err = curr.Decode(&user)
	fmt.Println(err)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "document not found", http.StatusNotFound)
		}
		return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
	}

	result := struct {
		Status string
		Data   store.User
	}{
		Status: "sucess",
		Data:   user,
	}
	return util.ResponseHandler(w, result, http.StatusOK)
}

func (s *Server) UpdateUserByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(r.Context(), "coffeeshop", "users")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		fmt.Println(err)
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	payload := r.Context().Value(middleware.AuthPayloadKey{}).(*token.Payload)
	if payload.Id != id.Hex() {
		return util.ResponseHandler(w, err, http.StatusForbidden)
	}

	err = r.ParseMultipartForm(int64(32 << 20))
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	fields := []string{"username", "phoneNumber"}
	data := bson.M{}
	for _, field := range fields {
		value := r.FormValue(field)
		if value != "" {
			data[field] = value
		}
	}

	if file, _, err := r.FormFile("avatar"); err == nil {
		resultChannel := make(chan imageResultParams, 2)
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			avatarFile, avatarName, err := util.ImageResizeProcessor(ctx, file)
			if err != nil {
				resultChannel <- imageResultParams{err: err}
				return
			}
			log.Print("goroutine 1 ", time.Now())
			resultChannel <- imageResultParams{avatarFile: avatarFile, avatarName: avatarName}
		}()

		wg.Add(1)
		avatarURL := make(chan string)
		go func(avatarData imageResultParams) {
			defer wg.Done()

			url, err := util.S3awsImageUpload(ctx, avatarData.avatarFile, "watamu-coffee-shop", avatarData.avatarName, "images/avatars")
			fmt.Println(url)
			if err != nil {
				resultChannel <- imageResultParams{err: err}
				return
			}
			avatarURL <- url
			close(avatarURL)
		}(<-resultChannel)

		select {
		case avatarName := <-avatarURL:
			data["avatar"] = avatarName

		case result := <-resultChannel:
			if result.err != nil {
				return util.ResponseHandler(w, err, http.StatusInternalServerError)
			}
		}
	}

	data["updated_at"] = time.Now()
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.M{"$set": data}

	var updatedDocument store.User
	err = collection.FindOneAndUpdate(ctx, filter, update).Decode(&updatedDocument)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "document not found", http.StatusNotFound)
		}
		return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
	}

	result := struct {
		Status string
	}{
		Status: "success",
	}
	return util.ResponseHandler(w, result, http.StatusOK)
}

func (s *Server) DeleteUserByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "users")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	payload := r.Context().Value(middleware.AuthPayloadKey{}).(*token.Payload)
	if payload.Id != id.Hex() {
		return util.ResponseHandler(w, err, http.StatusForbidden)
	}

	var deletedDocument bson.M
	err = collection.FindOneAndDelete(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&deletedDocument)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "invalid operation on data", http.StatusNotFound)
		}
		return util.ResponseHandler(w, "internal server error", http.StatusInternalServerError)
	}

	return util.ResponseHandler(w, "", http.StatusNoContent)
}

func (s *Server) ForgotPasswordHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "users")

	var forgotPassword forgotPasswordParams
	forgotPasswordBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}
	err = json.Unmarshal(forgotPasswordBytes, &forgotPassword)
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	var user store.User
	curr := collection.FindOne(ctx, bson.D{{Key: "email", Value: forgotPassword.Email}})
	err = curr.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, err.Error(), http.StatusNotFound)
		}

		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	timestamp := util.ResetToken(60)
	url := fmt.Sprintf("http://localhost:3000/resetpassword?token=%s&timestamp=%d", user.Id.Hex(), timestamp)

	result := struct {
		Status  string
		Message string
		Data    string
	}{
		Status:  "success",
		Message: "URL to reset your password sent to your email",
		Data:    url,
	}

	return util.ResponseHandler(w, result, http.StatusOK)
}

func (s *Server) ResetPasswordHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "users")

	queries := r.URL.Query()
	token := queries["token"][0]
	timestampStr := queries["timestamp"][0]

	timestamp, err := strconv.Atoi(timestampStr)
	if err != nil {
		err = fmt.Errorf("invalid URL reset timestamp %w", err)
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	isURLValid := time.Now().After(time.UnixMilli(int64(timestamp)))
	if isURLValid {
		fmt.Println(isURLValid)
		err = fmt.Errorf("expired URL reset token, kindly request for a new password reset token")
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	id, err := primitive.ObjectIDFromHex(token)
	if err != nil {
		err = fmt.Errorf("invalid URL reset token %w", err)
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	var passwordRest passwordResetParams
	passwordRestBytes, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("invalid data for password rset %w", err)
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	err = json.Unmarshal(passwordRestBytes, &passwordRest)
	if err != nil {
		err = fmt.Errorf("invalid data for password rset %w", err)
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	password := util.PasswordEncryption([]byte(passwordRest.Password))
	updatedAt := time.Now()
	passwordChangedAt := time.Now()

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "password", Value: password}, {Key: "updated_at", Value: updatedAt}, {Key: "password_changed_at", Value: passwordChangedAt}}}}

	var user store.User
	curr := collection.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: id}}, update)
	err = curr.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, err, http.StatusNotFound)
		}
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	result := struct {
		Status string
	}{
		Status: "success",
	}
	return util.ResponseHandler(w, result, http.StatusPermanentRedirect)
}

func (s *Server) VerifyAccountHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	queries := r.URL.Query()
	token := queries["token"][0]
	timestamp := queries["timestamp"][0]

	id, err := primitive.ObjectIDFromHex(token)
	if err != nil {
		err = fmt.Errorf("invalid account id %w", err)
		return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
	}

	mill, err := strconv.Atoi(timestamp)
	if err != nil {
		err = fmt.Errorf("invalid timestamp %w", err)
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	if time.Now().After(time.UnixMilli(int64(mill))) {
		err = fmt.Errorf("invalid timestamp expired %w", err)
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	var user store.User
	collection := s.Store.Collection(ctx, "coffeeshop", "users")
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "verified", Value: true}, {Key: "updated_at", Value: time.Now()}}}}
	curr := collection.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: id}}, update)
	err = curr.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = fmt.Errorf("document not found %w", err)
			return util.ResponseHandler(w, err.Error(), http.StatusNotFound)
		}
		return util.ResponseHandler(w, "", http.StatusInternalServerError)
	}

	result := struct {
		Status  string
		Message string
	}{
		Status:  "success",
		Message: "account verified",
	}
	return util.ResponseHandler(w, result, http.StatusOK)
}

func userRoutes(gmux *mux.Router, srv *Server) {
	getUsersRouter := gmux.Methods(http.MethodGet).Subrouter()
	postUserRouter := gmux.Methods(http.MethodPost).Subrouter()
	forgotPasswordRouter := gmux.Methods(http.MethodPost).Subrouter()
	updateUserRouter := gmux.Methods(http.MethodPut).Subrouter()
	resetPasswordRouter := gmux.Methods(http.MethodPut).Subrouter()
	deleteUserRouter := gmux.Methods(http.MethodDelete).Subrouter()

	getUsersRouter.Use(middleware.AuthMiddleware(srv.token))

	getAllUsersRouter := getUsersRouter.PathPrefix("/").Subrouter()
	getAllUsersRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin"))
	getAllUsersRouter.HandleFunc("/users", util.HandleFuncDecorator(srv.GetAllUsersHandlers))

	getUserByIdRouter := getUsersRouter.PathPrefix("/").Subrouter()
	getUserByIdRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin", "user"))
	getUserByIdRouter.HandleFunc("/users/{id}", util.HandleFuncDecorator(srv.GetUserByIdHandler))

	postUserRouter.HandleFunc("/signup", util.HandleFuncDecorator(srv.CreateUserHandler))
	postUserRouter.HandleFunc("/login", util.HandleFuncDecorator(srv.LoginUserHandler))

	updateUserRouter.Use(middleware.AuthMiddleware(srv.token))
	updateUserRouter.HandleFunc("/users/{id}", util.HandleFuncDecorator(srv.UpdateUserByIdHandler))

	deleteUserRouter.Use(middleware.AuthMiddleware(srv.token))
	deleteUserRouter.HandleFunc("/users/{id}", util.HandleFuncDecorator(srv.DeleteUserByIdHandler))
	forgotPasswordRouter.HandleFunc("/forgotpassword", util.HandleFuncDecorator(srv.ForgotPasswordHandler))
	resetPasswordRouter.HandleFunc("/resetpassword", util.HandleFuncDecorator(srv.ResetPasswordHandler))
}
