package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/silaselisha/coffee-api/pkg/store"
	"github.com/silaselisha/coffee-api/pkg/util"
	"github.com/stretchr/testify/require"
)

var user = util.CreateNewUser()

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
				testToken = result.Token
				userId = result.Data.Id.Hex()
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server := NewServer(ctx, mongoClient)
			url := "/users/signup"

			body, err := json.Marshal(test.body)
			require.NoError(t, err)
			require.NotEmpty(t, body)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			mux := server.(*Server)
			mux.Router.ServeHTTP(recorder, request)
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
			name: "login 200 status code",
			body: map[string]interface{}{
				"email":    user.Email,
				"password": user.Password,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				require.NotEmpty(t, data)

				var res []struct {
					Name  string
					Token string
				}
				err = json.Unmarshal(data, &res)
				require.NoError(t, err)
				log.Print(res)
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			server := NewServer(ctx, mongoClient)
			url := "/users/login"

			userCred, err := json.Marshal(test.body)
			require.NoError(t, err)
			require.NotEmpty(t, userCred)

			request := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(userCred))
			recorder := httptest.NewRecorder()
			server.(*Server).Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}
