package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/silaselisha/coffee-api/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	user := util.CreateNewUser()
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
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(mongoClient)
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
