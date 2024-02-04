package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateProduct(t *testing.T) {

	var tests = []struct {
		name       string
		bodyWriter func() (*bytes.Buffer, *multipart.Writer)
		check      func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "create a product",
			bodyWriter: func() (*bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				
				writer.WriteField("name", product.Name)
				writer.WriteField("price", strconv.FormatFloat(product.Price, 'E', -1, 64))
				writer.WriteField("summary", product.Summary)
				writer.WriteField("category", product.Category)
				writer.WriteField("description", product.Description)
				writer.WriteField("ingridients", strings.Join(product.Ingridients, " "))
				writer.WriteField("created_at", fmt.Sprint(product.CreatedAt))
				writer.WriteField("updated_at", fmt.Sprint(product.UpdatedAt))
				writer.Close()
				return body, writer
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				dataBytes, err := io.ReadAll(recorder.Body)
				require.Equal(t, nil, err)
				var productId string
				err = json.Unmarshal(dataBytes, &productId)
				require.Equal(t, nil, err)
				id = productId
				require.NotEmpty(t, id)
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			server := NewServer(testMonogoStore)

			url := "/products"
			recorder := httptest.NewRecorder()
			body, writer := test.bodyWriter()
			request, err := http.NewRequest(http.MethodPost, url, body)
			request.Header.Set("Content-Type", "multipart/form-data; boundary="+writer.Boundary())
			require.NoError(t, err)

			mux, ok := server.(*Server)
			require.Equal(t, true, ok)

			mux.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}
func TestUpdateProduct(t *testing.T) {
	var tests = []struct {
		name  string
		body  map[string]interface{}
		id    string
		check func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "update a product",
			id:   id,
			body: map[string]interface{}{
				"price": product.Price,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "update product by invalid id",
			id:   "65bcc06cbc92379c5b6fe79b",
			body: map[string]interface{}{
				"price": product.Price,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataBytes, err := json.Marshal(test.body)
			require.NoError(t, err)

			server := NewServer(testMonogoStore)

			url := fmt.Sprintf("/products/%s", test.id)
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(dataBytes))
			require.NoError(t, err)

			mux, ok := server.(*Server)
			require.Equal(t, true, ok)

			mux.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}

func TestGetAllProduct(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "get all products",
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(testMonogoStore)

			url := "/products"
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			mux, ok := server.(*Server)
			require.Equal(t, true, ok)

			mux.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}
func TestGetProduct(t *testing.T) {
	tests := []struct {
		name     string
		category map[string]interface{}
		id       string
		check    func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "get product by category & id",
			category: map[string]interface{}{"category": "beverages"},
			id:       id,
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:     "get product by category & invalid id",
			category: map[string]interface{}{"category": "beverages"},
			id:       "65bcc06cbc92379c5b6fe79b",
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:     "get product by invalid category & id",
			category: map[string]interface{}{"category": "beverage"},
			id:       id,
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(testMonogoStore)

			url := fmt.Sprintf("/products/%s/%s", test.category["category"], test.id)
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			mux, ok := server.(*Server)
			require.Equal(t, true, ok)

			mux.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		check func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "delete product by id",
			id:   id,
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNoContent, recorder.Code)
			},
		},
		{
			name: "delete product by invalid id",
			id:   "65bcc06cbc92379c5b6fe79b",
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := NewServer(testMonogoStore)

			url := fmt.Sprintf("/products/%s", test.id)
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			mux, ok := server.(*Server)
			require.Equal(t, true, ok)

			mux.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}
