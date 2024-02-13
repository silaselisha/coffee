// User services
// implementation SSO, socials signup & signin and traditional
// user signup & login
// cookie session management and refresh tokens
// password reset functionality; forgot apssword?
// account update info functionality [Avatar, PhoneNumber, Password]
// deactivate account and set for permanet deletion 30day without login
// delete account with all it's associated data
package handler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/token"
	"github.com/silaselisha/coffee-api/pkg/util"
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
	collection := s.db.Collection(ctx, "coffeeshop", "users")
	curr := collection.FindOne(ctx, bson.D{{Key: "email", Value: credentials.Email}})
	if err := curr.Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "document not found", http.StatusBadRequest)
		}
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	if !util.ComparePasswordEncryption(credentials.Password, user.Password) {
		return util.ResponseHandler(w, "invalid email or password", http.StatusBadRequest)
	}

	jwtToken := token.NewToken(s.envs.SecretAccessKey)
	token, err := jwtToken.CreateToken(ctx, user.Email, time.Second*90)
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
	collection := s.db.Collection(ctx, "coffeeshop", "users")
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}, {Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
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

	jwtoken := token.NewToken(s.envs.SecretAccessKey)
	minutes, err := strconv.Atoi(s.envs.JwtExpiresAt)
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusInternalServerError)
	}

	duration := time.Minute * time.Duration(minutes)
	token, err := jwtoken.CreateToken(ctx, data.Email, duration)
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
	collection := s.db.Collection(ctx, "coffeeshop", "users")

	var users store.UserList
	cur, err := collection.Find(ctx, bson.D{{}})
	if err != nil {
		log.Print(err)
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var user store.User
		err := cur.Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				break
			}
			log.Print(err)
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
	collection := s.db.Collection(ctx, "coffeeshop", "users")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	result := collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}})
	var user store.User
	err = result.Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "document not found", http.StatusNotFound)
		}
		return util.ResponseHandler(w, err, http.StatusInternalServerError)
	}

	res := struct {
		Status string
		Data   store.User
	}{
		Status: "sucess",
		Data:   user,
	}
	return util.ResponseHandler(w, res, http.StatusOK)
}

func (s *Server) UpdateUserByIdHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	collection := s.db.Collection(ctx, "coffeeshop", "users")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		return util.ResponseHandler(w, err, http.StatusBadRequest)
	}

	err = r.ParseMultipartForm(int64(32 << 20))
	if err != nil {
		log.Print(err.Error())
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
		avatarName := make(chan string)
		avatarFile := make(chan []byte)
		errs := make(chan error)
		go func() {
			avatar, fileName, err := util.ImageResizeProcessor(ctx, file)
			if err != nil {
				errs <- err
				return
			}

			err = util.S3awsImageUpload(ctx, avatar, "watamu-coffee-shop", fileName)
			if err != nil {
				errs <- err
				return
			}

			avatarName <- fileName
			avatarFile <- avatar	
			close(avatarName)
			close(avatarFile)
		}()

		select {
		case avatar := <-avatarName:
			data["avatar"] = avatar
		case err := <-errs:
			if err != nil {
				return util.ResponseHandler(w, err, http.StatusBadRequest)
			}
		}
	}

	data["updated_at"] = time.Now()
	log.Print(data)
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.M{"$set": data}

	var updatedDocument store.User
	err = collection.FindOneAndUpdate(ctx, filter, update).Decode(&updatedDocument)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Print(err.Error())
			return util.ResponseHandler(w, err, http.StatusBadRequest)
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
	collection := s.db.Collection(ctx, "coffeeshop", "users")

	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		return util.ResponseHandler(w, err.Error(), http.StatusBadRequest)
	}

	var deletedDocument bson.M
	err = collection.FindOneAndDelete(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&deletedDocument)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return util.ResponseHandler(w, "invalid operation on data", http.StatusBadRequest)
		}
		return util.ResponseHandler(w, "internal server error", http.StatusInternalServerError)
	}
	return util.ResponseHandler(w, "", http.StatusNoContent)
}

func userRoutes(gmux *mux.Router, srv *Server) {
	getUserRouter := gmux.Methods(http.MethodGet).Subrouter()
	postUserRouter := gmux.Methods(http.MethodPost).Subrouter()
	updateUserRouter := gmux.Methods(http.MethodPut).Subrouter()

	getUserRouter.HandleFunc("/users", util.HandleFuncDecorator(srv.GetAllUsersHandlers))
	postUserRouter.HandleFunc("/users/signup", util.HandleFuncDecorator(srv.CreateUserHandler))
	postUserRouter.HandleFunc("/users/login", util.HandleFuncDecorator(srv.LoginUserHandler))
	updateUserRouter.HandleFunc("/users/{id}", util.HandleFuncDecorator(srv.UpdateUserByIdHandler))
}
