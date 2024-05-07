// delete account and it's s3 object
// updating user profile image then delete the exisitng one
// resize the image (research)
// email user code URL for verification, reset password
// create a path to verify phone number
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/hibiken/asynq"
	"github.com/silaselisha/coffee-api/internal"
	"github.com/silaselisha/coffee-api/pkg/middleware"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/token"
	"github.com/silaselisha/coffee-api/types"
	"github.com/silaselisha/coffee-api/workers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Server) LoginUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	credentialsBytes, err := io.ReadAll(r.Body)
	if err != nil {
		response := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, response, http.StatusBadRequest)
	}

	var credentials types.UserLoginParams
	json.Unmarshal(credentialsBytes, &credentials)
	err = s.vd.Struct(credentials)
	if err != nil {
		response := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, response, http.StatusBadRequest)
	}

	var user store.User
	collection := s.Store.Collection(ctx, "coffeeshop", "users")
	curr := collection.FindOne(ctx, bson.D{{Key: "email", Value: credentials.Email}})
	if err := curr.Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			response := internal.NewErrorResponse("failed", err.Error())
			return internal.ResponseHandler(w, response, http.StatusNotFound)
		}

		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	if !internal.ComparePasswordEncryption(credentials.Password, user.Password) {
		err := errors.New("invalid user password or email address")
		response := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, response, http.StatusBadRequest)
	}

	jwtToken := token.NewToken(s.envs.SECRET_ACCESS_KEY)
	days, err := strconv.Atoi(s.envs.JWT_EXPIRES_AT)
	if err != nil {
		return internal.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	hrs := fmt.Sprintf("%dh", (days * 24))
	duration, err := time.ParseDuration(hrs)
	if err != nil {
		response := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, response, http.StatusInternalServerError)
	}

	token, err := jwtToken.CreateToken(ctx, duration, user.Id.Hex(), user.Email)
	if err != nil {
		response := internal.NewErrorResponse("failed", err.Error())
		return internal.ResponseHandler(w, response, http.StatusInternalServerError)
	}
	res := struct {
		Status string `json:"status"`
		Token  string `json:"token"`
	}{
		Status: "success",
		Token:  token,
	}
	return internal.ResponseHandler(w, res, http.StatusOK)
}

func (s *Server) CreateUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	session, err := s.Store.TxnStartSession(ctx)
	if err != nil {
		return session.AbortTransaction(ctx)
	}

	defer func() {
		if abortError := session.AbortTransaction(ctx); err != nil {
			err = abortError
		}
	}()

	defer session.EndSession(ctx)
	response, err := session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		collection := s.Store.Collection(ctx, "coffeeshop", "users")
		_, err = collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
			{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
		})

		if err != nil {
			return nil, err
		}

		var user store.User

		userBytes, err := io.ReadAll(r.Body)
		if err != nil {
			if err == io.EOF {
				return nil, err
			}
			return nil, err
		}

		err = json.Unmarshal(userBytes, &user)
		if err != nil {
			return nil, err
		}

		if err := s.vd.Struct(user); err != nil {
			return nil, err
		}

		hashedPassword := internal.PasswordEncryption([]byte(user.Password))
		user.Id = primitive.NewObjectID()
		user.Role = "user"
		user.Avatar = "default.jpeg"
		user.Password = hashedPassword
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		_, err = collection.InsertOne(ctx, user)
		if err != nil {
			return nil, err
		}

		opts := []asynq.Option{
			asynq.MaxRetry(3),
			asynq.ProcessIn(3 * time.Second),
			asynq.Queue(workers.CriticalQueue),
		}
		err = s.distributor.SendVerificationMailTask(ctx, &types.PayloadSendMail{Email: user.Email}, opts...)
		if err != nil {
			return nil, err
		}

		resposne := &types.UserResponseParams{
			Id:          user.Id.Hex(),
			Avatar:      user.Avatar,
			UserName:    user.UserName,
			Role:        user.Role,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Verified:    user.Verified,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}

		err = session.CommitTransaction(ctx)
		if err != nil {
			return nil, err
		}

		return resposne, nil
	}, &options.TransactionOptions{})

	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return internal.ResponseHandler(w, internal.NewErrorResponse("failed", fmt.Errorf("document not found %w", err).Error()), http.StatusNotFound)
		case errors.As(err, &mongo.WriteException{}):
			wrtExcp, _ := err.(mongo.WriteException)
			if wrtExcp.WriteErrors[0].Code == 11000 {
				return internal.ResponseHandler(w, internal.NewErrorResponse("failed", fmt.Errorf("document already exists %w", err).Error()), http.StatusBadRequest)
			}
		case errors.Is(err, &json.SyntaxError{}):
			return internal.ResponseHandler(w, internal.NewErrorResponse("failed", fmt.Errorf("invalid data input for operation %w", err).Error()), http.StatusBadRequest)
		case err.(validator.ValidationErrors) != nil:
			return internal.ResponseHandler(w, internal.NewErrorResponse("failed", fmt.Errorf("invalid data input for operation %w", err).Error()), http.StatusBadRequest)
		default:
			return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
		}
	}

	jwtoken := token.NewToken(s.envs.SECRET_ACCESS_KEY)
	days, err := strconv.Atoi(s.envs.JWT_EXPIRES_AT)
	if err != nil {
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	hrs := fmt.Sprintf("%dh", (days * 24))
	duration, err := time.ParseDuration(hrs)
	if err != nil {
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}
	user := response.(*types.UserResponseParams)
	token, err := jwtoken.CreateToken(ctx, duration, user.Id, user.Email)
	if err != nil {
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	result := struct {
		Status string                    `json:"status"`
		Token  string                    `json:"token"`
		Data   *types.UserResponseParams `json:"data"`
	}{
		Status: "success",
		Token:  token,
		Data:   user,
	}
	return internal.ResponseHandler(w, result, http.StatusCreated)
}

func (s *Server) GetAllUsersHandlers(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "users")

	var users types.UserResponseListParams
	curr, err := collection.Find(ctx, bson.D{{}})
	if err != nil {
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	defer curr.Close(ctx)
	for curr.Next(ctx) {
		var user store.User
		err := curr.Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				break
			}

			return internal.ResponseHandler(w, err, http.StatusInternalServerError)
		}

		users = append(users, types.UserResponseParams{
			Id:          user.Id.Hex(),
			Avatar:      user.Avatar,
			UserName:    user.UserName,
			Role:        user.Role,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Verified:    user.Verified,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		})
	}

	result := struct {
		Status string                       `json:"status"`
		Result int32                        `json:"result"`
		Data   types.UserResponseListParams `json:"data"`
	}{
		Status: "success",
		Result: int32(len(users)),
		Data:   users,
	}
	return internal.ResponseHandler(w, result, http.StatusOK)
}

func (s *Server) GetUserByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "users")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}

	payload := ctx.Value(types.AuthPayloadKey{}).(*token.Payload)
	userInfo := ctx.Value(types.AuthUserInfoKey{}).(*types.UserInfo)

	if payload.Id != id.Hex() && userInfo.Role != "admin" {
		err := errors.New("user only allowed to retrive their person account")
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusForbidden)
	}

	curr := collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}})
	var user store.User
	err = curr.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return internal.ResponseHandler(w, internal.NewErrorResponse("failed", fmt.Errorf("document not found %w", err).Error()), http.StatusNotFound)
		}

		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	resposne := types.UserResponseParams{
		Id:          user.Id.Hex(),
		Avatar:      user.Avatar,
		UserName:    user.UserName,
		Role:        user.Role,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Verified:    user.Verified,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	result := struct {
		Status string                   `json:"status"`
		Data   types.UserResponseParams `json:"data"`
	}{
		Status: "success",
		Data:   resposne,
	}
	return internal.ResponseHandler(w, result, http.StatusOK)
}

func (s *Server) UpdateUserByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(r.Context(), "coffeeshop", "users")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}

	userInfo := r.Context().Value(types.AuthUserInfoKey{}).(*types.UserInfo)
	if id.Hex() != userInfo.Id.Hex() {
		err := errors.New("user only allowed to retrive their person account")
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusForbidden)
	}

	err = r.ParseMultipartForm(int64(32 << 20))
	if err != nil {
		return internal.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	fields := []string{"username", "phoneNumber"}
	data := bson.M{}
	for _, field := range fields {
		value := r.FormValue(field)
		if value != "" {
			data[field] = value
		}
	}

	errs := make(chan error)
	fileName := make(chan string)
	if file, _, err := r.FormFile("avatar"); err == nil {
		go func() {
			defer file.Close()
			data, filename, extension, err := internal.ImageProcessor(ctx, file, &types.FileMetadata{ContetntType: "image"})
			if err != nil {
				errs <- err
				return
			}

			objectKey := fmt.Sprintf("images/avatars/%s", filename)
			err = s.S3Client.UploadImage(ctx, objectKey, s.envs.S3_BUCKET_NAME, extension, data)
			if err != nil {
				errs <- err
				return
			}

			err = s.distributor.SendS3ObjectUploadTask(ctx, &types.PayloadUploadImage{
				Image:     data,
				ObjectKey: objectKey,
				Extension: extension,
			}, []asynq.Option{asynq.ProcessIn(1 * time.Second),
				asynq.MaxRetry(3),
				asynq.Queue(workers.CriticalQueue)}...)

			if err != nil {
				errs <- err
			}
			err = s.distributor.SendS3ObjectDeleteTask(ctx, []string{userInfo.Avatar}, []asynq.Option{asynq.ProcessIn(3 * time.Minute),
				asynq.MaxRetry(3),
				asynq.Queue(workers.CriticalQueue)}...)
			if err != nil {
				errs <- err
				return
			}

			fileName <- objectKey
			close(errs)
			close(fileName)
		}()

		select {
		case filename, ok := <-fileName:
			if !ok {
				return internal.ResponseHandler(w, internal.NewErrorResponse("failed", fmt.Errorf("image file name error").Error()), http.StatusInternalServerError)
			}
			data["avatar"] = filename
		case err := <-errs:
			if err != nil {
				return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
			}
		}
	}

	data["updated_at"] = time.Now()
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.M{"$set": data}

	newDocs := options.After
	var updatedDocument store.User
	err = collection.FindOneAndUpdate(ctx, filter, update, &options.FindOneAndUpdateOptions{
		ReturnDocument: &newDocs,
	}).Decode(&updatedDocument)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return internal.ResponseHandler(w, internal.NewErrorResponse("failed", fmt.Errorf("document not found %w", err).Error()), http.StatusNotFound)
		}

		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	updatedUser := types.UserResponseParams{
		Id:          updatedDocument.Id.Hex(),
		Avatar:      updatedDocument.Avatar,
		UserName:    updatedDocument.UserName,
		Role:        updatedDocument.Role,
		Email:       updatedDocument.Email,
		PhoneNumber: updatedDocument.PhoneNumber,
		Verified:    updatedDocument.Verified,
		CreatedAt:   updatedDocument.CreatedAt,
		UpdatedAt:   updatedDocument.UpdatedAt,
	}

	result := struct {
		Status string                   `json:"status"`
		Data   types.UserResponseParams `json:"data"`
	}{
		Status: "success",
		Data:   updatedUser,
	}
	return internal.ResponseHandler(w, result, http.StatusOK)
}

func (s *Server) DeleteUserByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "users")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}

	userInfo := r.Context().Value(types.AuthUserInfoKey{}).(*types.UserInfo)
	if userInfo.Id.Hex() != id.Hex() {
		err := errors.New("user only allowed to retrive their person account")
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusForbidden)
	}

	var user store.User
	err = collection.FindOneAndDelete(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return internal.ResponseHandler(w, internal.NewErrorResponse("failed", fmt.Errorf("document not found %w", err).Error()), http.StatusNotFound)
		}

		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	errs := make(chan error)
	go func() {
		avatarURL := user.Avatar
		err := s.S3Client.DeleteImage(ctx, avatarURL, s.envs.S3_BUCKET_NAME)
		if err != nil {
			errs <- err
			return
		}
		close(errs)
	}()

	err = <-errs
	if err != nil {
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	return internal.ResponseHandler(w, "", http.StatusNoContent)
}

func (s *Server) ForgotPasswordHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "users")

	var forgotPassword types.ForgotPasswordParams
	forgotPasswordBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}
	err = json.Unmarshal(forgotPasswordBytes, &forgotPassword)
	if err != nil {
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}

	var user store.User
	curr := collection.FindOne(ctx, bson.D{{Key: "email", Value: forgotPassword.Email}})
	err = curr.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return internal.ResponseHandler(w, internal.NewErrorResponse("failed", fmt.Errorf("document not found %w", err).Error()), http.StatusNotFound)
		}

		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	opts := []asynq.Option{
		asynq.ProcessIn(1 * time.Minute),
		asynq.MaxRetry(10),
		asynq.Queue("critical"),
	}
	err = s.distributor.SendPasswordResetMailTask(ctx, &types.PayloadSendMail{Email: user.Email}, opts...)
	if err != nil {
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	result := struct {
		Status string `json:"status"`
		Data   string `json:"data"`
	}{
		Status: "success",
		Data:   "URL to reset your password sent to your email",
	}

	return internal.ResponseHandler(w, result, http.StatusOK)
}

func (s *Server) ResetPasswordHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.Store.Collection(ctx, "coffeeshop", "users")

	queries := r.URL.Query()
	token := queries["token"][0]
	timestampStr := queries["timestamp"][0]

	timestamp, err := strconv.Atoi(timestampStr)
	if err != nil {
		err = fmt.Errorf("invalid URL reset timestamp %w", err)
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}

	isURLValid := time.Now().After(time.UnixMilli(int64(timestamp)))
	if isURLValid {

		err = fmt.Errorf("expired URL reset token, kindly request for a new password reset token")
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}

	id, err := primitive.ObjectIDFromHex(token)
	if err != nil {
		err = fmt.Errorf("invalid URL reset token %w", err)
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}

	var passwordRest types.PasswordResetParams
	passwordRestBytes, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("invalid data for password rset %w", err)
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}

	err = json.Unmarshal(passwordRestBytes, &passwordRest)
	if err != nil {
		err = fmt.Errorf("invalid data for password reset %w", err)
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}

	password := internal.PasswordEncryption([]byte(passwordRest.Password))
	updatedAt := time.Now()
	passwordChangedAt := time.Now()

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "password", Value: password}, {Key: "updated_at", Value: updatedAt}, {Key: "password_changed_at", Value: passwordChangedAt}}}}

	var user store.User
	curr := collection.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: id}}, update)
	err = curr.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return internal.ResponseHandler(w, internal.NewErrorResponse("failed", fmt.Errorf("document not found %w", err).Error()), http.StatusNotFound)
		}

		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	result := struct {
		Status string `json:"status"`
	}{
		Status: "success",
	}
	return internal.ResponseHandler(w, result, http.StatusPermanentRedirect)
}

func (s *Server) VerifyAccountHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	queries := r.URL.Query()
	token := queries["token"][0]
	timestamp := queries["timestamp"][0]

	id, err := primitive.ObjectIDFromHex(token)
	if err != nil {
		err = fmt.Errorf("invalid account id %w", err)
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	mill, err := strconv.Atoi(timestamp)
	if err != nil {
		err = fmt.Errorf("invalid timestamp %w", err)
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}

	if time.Now().After(time.UnixMilli(int64(mill))) {
		err = fmt.Errorf("invalid timestamp expired %w", err)
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusBadRequest)
	}

	var user store.User
	collection := s.Store.Collection(ctx, "coffeeshop", "users")
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "verified", Value: true}, {Key: "updated_at", Value: time.Now()}}}}
	curr := collection.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: id}}, update)
	err = curr.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = fmt.Errorf("document not found %w", err)
			return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusNotFound)
		}
		return internal.ResponseHandler(w, internal.NewErrorResponse("failed", err.Error()), http.StatusInternalServerError)
	}

	result := struct {
		Status string `json:"status"`
		Data   string `json:"data"`
	}{
		Status: "success",
		Data:   "account verified",
	}
	return internal.ResponseHandler(w, result, http.StatusOK)
}

func userRoutes(gmux *mux.Router, srv *Server) {
	userGetRouter := gmux.Methods(http.MethodGet).Subrouter()
	postUserRouter := gmux.Methods(http.MethodPost).Subrouter()
	forgotPasswordRouter := gmux.Methods(http.MethodPost).Subrouter()
	updateUserRouter := gmux.Methods(http.MethodPut).Subrouter()
	resetPasswordRouter := gmux.Methods(http.MethodPut).Subrouter()
	deleteUserRouter := gmux.Methods(http.MethodDelete).Subrouter()

	userGetRouter.Use(middleware.AuthMiddleware(srv.token))

	getAllUsersRouter := userGetRouter.PathPrefix("/").Subrouter()
	getAllUsersRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin"))
	getAllUsersRouter.HandleFunc("/users", internal.HandleFuncDecorator(srv.GetAllUsersHandlers))

	getUserByIdRouter := userGetRouter.PathPrefix("/").Subrouter()
	getUserByIdRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin", "user"))
	getUserByIdRouter.HandleFunc("/users/{id}", internal.HandleFuncDecorator(srv.GetUserByIdHandler))

	postUserRouter.HandleFunc("/signup", internal.HandleFuncDecorator(srv.CreateUserHandler))
	postUserRouter.HandleFunc("/login", internal.HandleFuncDecorator(srv.LoginUserHandler))

	updateUserRouter.Use(middleware.AuthMiddleware(srv.token))
	updateUserRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin", "user"))
	updateUserRouter.HandleFunc("/users/{id}", internal.HandleFuncDecorator(srv.UpdateUserByIdHandler))

	deleteUserRouter.Use(middleware.AuthMiddleware(srv.token))
	deleteUserRouter.Use(middleware.RestrictToMiddleware(srv.Store, "admin", "user"))
	deleteUserRouter.HandleFunc("/users/{id}", internal.HandleFuncDecorator(srv.DeleteUserByIdHandler))
	forgotPasswordRouter.HandleFunc("/forgotpassword", internal.HandleFuncDecorator(srv.ForgotPasswordHandler))
	resetPasswordRouter.HandleFunc("/resetpassword", internal.HandleFuncDecorator(srv.ResetPasswordHandler))
}
