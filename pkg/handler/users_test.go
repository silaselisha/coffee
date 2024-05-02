package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/util"
	"github.com/stretchr/testify/require"
)

var user = util.CreateNewUser("johndoe@test.com", "doe", "+1(571)360-6677", "user")

func TestCreateUserSignup(t *testing.T) {
	tests := []struct {
		name  string
		body  map[string]interface{}
		check func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "user signup 201 status code",
			body: map[string]interface{}{
				"username":    user.UserName,
				"email":       user.Email,
				"password":    user.Password,
				"phoneNumber": user.PhoneNumber,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var result struct {
					Status string
					Token  string
					Data   store.User
				}
				body, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				require.NotEmpty(t, body)

				err = json.Unmarshal(body, &result)
				require.NoError(t, err)
				userTestToken = result.Token
				userID = result.Data.Id.Hex()
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "user signup 400 status code",
			body: map[string]interface{}{
				"username": user.UserName,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "user signup 400 status code",
			body: map[string]interface{}{
				"email": user.Email,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "user signup 400 status code",
			body: map[string]interface{}{
				"email":    user.Email,
				"username": user.UserName,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "user signup 400 status code",
			body: map[string]interface{}{},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			url := "/signup"

			body, err := json.Marshal(test.body)
			require.NoError(t, err)
			require.NotEmpty(t, body)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))

			querier := NewServer(ctx, mongoClient, distributor)
			server, ok := querier.(*Server)
			require.Equal(t, true, ok)

			server.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}

func TestUserLogin(t *testing.T) {
	tests := []struct {
		name  string
		body  map[string]interface{}
		check func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "user login | 200 status code",
			body: map[string]interface{}{
				"email":    user.Email,
				"password": user.Password,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				require.NotEmpty(t, data)

				var res struct {
					Status string
					Token  string
				}
				err = json.Unmarshal(data, &res)
				require.NoError(t, err)
				userTestToken = res.Token
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "admin login | 200 status code",
			body: map[string]interface{}{
				"email":    "admin@aws.ac.uk",
				"password": "Abstract$87",
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				require.NotEmpty(t, data)

				var res struct {
					Status string
					Token  string
				}
				err = json.Unmarshal(data, &res)
				require.NoError(t, err)
				adminTestToken = res.Token
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "login user wrong email | 404 status code",
			body: map[string]interface{}{
				"email":    "test@test.com",
				"password": user.Password,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "login user wrong password | 400 status code",
			body: map[string]interface{}{
				"email":    user.Email,
				"password": "abstract&87",
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "login user invalid credentials | 400 status code",
			body: map[string]interface{}{},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			url := "/login"

			userCred, err := json.Marshal(test.body)
			require.NoError(t, err)
			require.NotEmpty(t, userCred)

			request := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(userCred))
			recorder := httptest.NewRecorder()

			querier := NewServer(ctx, mongoClient, distributor)
			server, ok := querier.(*Server)
			require.Equal(t, true, ok)

			server.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}

func TestGetAllUsers(t *testing.T) {
	tests := []struct {
		name  string
		token string
		body  map[string]interface{}
		check func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "get all users | status 200",
			token: adminTestToken,
			body:  map[string]interface{}{},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			url := "/users"
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, url, nil)
			request.Header.Set("authorization", fmt.Sprintf("Bearer %s", test.token))

			querier := NewServer(ctx, mongoClient, distributor)
			server, ok := querier.(*Server)
			require.Equal(t, true, ok)

			server.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}

func TestGetUserById(t *testing.T) {
	tests := []struct {
		name  string
		body  map[string]interface{}
		id    string
		token string
		check func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "get user by id | status code 200",
			body:  map[string]interface{}{},
			id:    userID,
			token: userTestToken,
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "get user by id | status code 200",
			body:  map[string]interface{}{},
			id:    adminID,
			token: adminTestToken,
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "get user by id | status code 400",
			body:  map[string]interface{}{},
			id:    "1234",
			token: userTestToken,
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "get user by id | status code 403",
			body:  map[string]interface{}{},
			id:    "65bcc06cbc92379c5b6fe79b",
			token: userTestToken,
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:  "get user by id | status code 403",
			body:  map[string]interface{}{},
			id:    userID,
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6ImFsM3hhQGF3cy5hYy51ayIsIklkIjoiNjVkMWYzYzRkZjRlNjM4NjAxYTczNjliIiwiSXNzdWVkQXQiOiIyMDI0LTAyLTE4VDE1OjEwOjQ0LjgzNjEyNjE4NiswMzowMCIsIkV4cGlyZWRBdCI6IjIwMjQtMDUtMThUMTU6MTA6NDQuODM2MTI2MjQ4KzAzOjAwIn0.P26Jmris4dfH4v-sayNmnFty8yEtOXGhqb4xgtlXkPk",
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:  "get user by id | status code 404",
			body:  map[string]interface{}{},
			id:    "65d1f3c4df4e638601a7369b",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6ImFsM3hhQGF3cy5hYy51ayIsIklkIjoiNjVkMWYzYzRkZjRlNjM4NjAxYTczNjliIiwiSXNzdWVkQXQiOiIyMDI0LTAyLTE4VDE1OjEwOjQ0LjgzNjEyNjE4NiswMzowMCIsIkV4cGlyZWRBdCI6IjIwMjQtMDUtMThUMTU6MTA6NDQuODM2MTI2MjQ4KzAzOjAwIn0.P26Jmris4dfH4v-sayNmnFty8yEtOXGhqb4xgtlXkPk",
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				fmt.Println(recorder.Code)
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			url := fmt.Sprintf("/users/%s", test.id)
			request := httptest.NewRequest(http.MethodGet, url, nil)
			recorder := httptest.NewRecorder()
			request.Header.Set("authorization", fmt.Sprintf("Bearer %s", test.token))

			querier := NewServer(ctx, mongoClient, distributor)
			server, ok := querier.(*Server)
			require.Equal(t, true, ok)

			fmt.Println(url)
			server.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}

func TestUpdateUserById(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		token      string
		bodyWriter func() (*bytes.Buffer, *multipart.Writer)
		check      func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "update user by id | status code 200",
			id:    userID,
			token: userTestToken,
			bodyWriter: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				writer.WriteField("username", "alena")
				writer.WriteField("phoneNumber", "+1 312 484 4884")
				defer writer.Close()
				return body, writer
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "update user's avatar | status code 200",
			id:    userID,
			token: userTestToken,
			bodyWriter: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				palette := []color.Color{color.Black, color.White}
				w, err := writer.CreateFormFile("avatar", "testprofileimage.jpg")
				require.NoError(t, err)
				img := image.NewPaletted(image.Rect(0, 0, 800, 400), palette)
				err = png.Encode(w, img)
				require.NoError(t, err)

				defer writer.Close()
				return body, writer
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:  "update user by id | status code 403",
			id:    userID,
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6ImFsM3hhQGF3cy5hYy51ayIsIklkIjoiNjVkMWYzYzRkZjRlNjM4NjAxYTczNjliIiwiSXNzdWVkQXQiOiIyMDI0LTAyLTE4VDE1OjEwOjQ0LjgzNjEyNjE4NiswMzowMCIsIkV4cGlyZWRBdCI6IjIwMjQtMDUtMThUMTU6MTA6NDQuODM2MTI2MjQ4KzAzOjAwIn0.P26Jmris4dfH4v-sayNmnFty8yEtOXGhqb4xgtlXkPk",
			bodyWriter: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				defer writer.Close()
				return body, writer
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:  "update user by id | status code 403",
			id:    "65d1f3c4df4e638601a7369b",
			token: userTestToken,
			bodyWriter: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				defer writer.Close()
				return body, writer
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:  "update user by id | status code 400",
			id:    "1234",
			token: userTestToken,
			bodyWriter: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				defer writer.Close()
				return body, writer
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:  "update user by id | status code 404",
			id:    "65d1f3c4df4e638601a7369b",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6ImFsM3hhQGF3cy5hYy51ayIsIklkIjoiNjVkMWYzYzRkZjRlNjM4NjAxYTczNjliIiwiSXNzdWVkQXQiOiIyMDI0LTAyLTE4VDE1OjEwOjQ0LjgzNjEyNjE4NiswMzowMCIsIkV4cGlyZWRBdCI6IjIwMjQtMDUtMThUMTU6MTA6NDQuODM2MTI2MjQ4KzAzOjAwIn0.P26Jmris4dfH4v-sayNmnFty8yEtOXGhqb4xgtlXkPk",
			bodyWriter: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				defer writer.Close()
				return body, writer
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		url := fmt.Sprintf("/users/%s", test.id)
		body, writer := test.bodyWriter()
		recorder := httptest.NewRecorder()
		request, err := http.NewRequest(http.MethodPut, url, body)
		require.NoError(t, err)
		request.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
		request.Header.Set("authorization", fmt.Sprintf("Bearer %s", test.token))

		querier := NewServer(ctx, mongoClient, distributor)
		server, ok := querier.(*Server)
		require.Equal(t, true, ok)

		server.Router.ServeHTTP(recorder, request)
		test.check(t, recorder)
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name   string
		body   map[string]interface{}
		userId string
		token  string
		check  func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{

		{
			name:   "delete user's account | status 204",
			body:   map[string]interface{}{},
			userId: userID,
			token:  userTestToken,
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name:   "delete user's account | status 400",
			body:   map[string]interface{}{},
			userId: "1234",
			token:  userTestToken,
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "delete user's account | status 403",
			body:   map[string]interface{}{},
			userId: "65bcc06cbc92379c5b6fe79b",
			token:  userTestToken,
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:   "delete user's account | status 404",
			body:   map[string]interface{}{},
			userId: "65d1f3c4df4e638601a7369b",
			token:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6ImFsM3hhQGF3cy5hYy51ayIsIklkIjoiNjVkMWYzYzRkZjRlNjM4NjAxYTczNjliIiwiSXNzdWVkQXQiOiIyMDI0LTAyLTE4VDE1OjEwOjQ0LjgzNjEyNjE4NiswMzowMCIsIkV4cGlyZWRBdCI6IjIwMjQtMDUtMThUMTU6MTA6NDQuODM2MTI2MjQ4KzAzOjAwIn0.P26Jmris4dfH4v-sayNmnFty8yEtOXGhqb4xgtlXkPk",
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			url := fmt.Sprintf("/users/%s", test.userId)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodDelete, url, nil)
			request.Header.Set("authorization", fmt.Sprintf("Bearer %s", test.token))

			querier := NewServer(ctx, mongoClient, distributor)
			server, ok := querier.(*Server)
			require.Equal(t, true, ok)

			server.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}
