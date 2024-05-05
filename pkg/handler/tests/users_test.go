package handler__test

import (
	"bytes"
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

	"github.com/silaselisha/coffee-api/internal"
	"github.com/silaselisha/coffee-api/types"
	"github.com/stretchr/testify/require"
)

var user = internal.CreateNewUser("johndoe@test.com", "doe", "+1(571)360-6677", "user")

func TestCreateUserSignup(t *testing.T) {
	testCases := []struct {
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
					Data   types.UserResponseParams
				}
				body, err := io.ReadAll(recorder.Body)

				require.NoError(t, err)
				require.NotEmpty(t, body)

				err = json.Unmarshal(body, &result)
				require.NoError(t, err)
				userTestToken = result.Token
				userID = result.Data.Id
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "user signup 400 status code",
			body: map[string]interface{}{
				"username":    user.UserName,
				"email":       "john@gmail.com",
				"password":    user.Password,
				"phoneNumber": user.PhoneNumber,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "user signup 400 status code",
			body: map[string]interface{}{
				"username":    "jo3",
				"email":       user.Email,
				"password":    user.Password,
				"phoneNumber": user.PhoneNumber,
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			url := "/api/v1/signup"
			body, err := json.Marshal(tc.body)
			require.NoError(t, err)
			require.NotEmpty(t, body)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))

			server.Router.ServeHTTP(recorder, request)
			tc.check(t, recorder)
		})
	}
}

func TestUserLogin(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			url := "/api/v1/login"
			userCred, err := json.Marshal(tc.body)
			require.NoError(t, err)
			require.NotEmpty(t, userCred)

			request := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(userCred))
			recorder := httptest.NewRecorder()

			server.Router.ServeHTTP(recorder, request)
			tc.check(t, recorder)
		})
	}
}

func TestGetAllUsers(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			url := "/api/v1/users"
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, url, nil)
			request.Header.Set("authorization", fmt.Sprintf("Bearer %s", tc.token))

			server.Router.ServeHTTP(recorder, request)
			tc.check(t, recorder)
		})
	}
}

func TestGetUserById(t *testing.T) {
	testCases := []struct {
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
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			url := fmt.Sprintf("/api/v1/users/%s", tc.id)
			request := httptest.NewRequest(http.MethodGet, url, nil)
			recorder := httptest.NewRecorder()
			request.Header.Set("authorization", fmt.Sprintf("Bearer %s", tc.token))

			server.Router.ServeHTTP(recorder, request)
			tc.check(t, recorder)
		})
	}
}

func TestUpdateUserById(t *testing.T) {
	testCases := []struct {
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
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			url := fmt.Sprintf("/api/v1/users/%s", tc.id)
			body, writer := tc.bodyWriter()
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodPut, url, body)
			require.NoError(t, err)
			request.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
			request.Header.Set("authorization", fmt.Sprintf("Bearer %s", tc.token))

			server.Router.ServeHTTP(recorder, request)
			tc.check(t, recorder)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	testCases := []struct {
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
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			url := fmt.Sprintf("/api/v1/users/%s", tc.userId)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodDelete, url, nil)
			request.Header.Set("authorization", fmt.Sprintf("Bearer %s", tc.token))

			server.Router.ServeHTTP(recorder, request)
			tc.check(t, recorder)
		})
	}
}
