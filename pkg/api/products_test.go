package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateProducts(t *testing.T) {
	var tests = []struct {
		name  string
		body  map[string]interface{}
		check func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "create a product",
			body: map[string]interface{}{
				"name":        product.Name,
				"price":       product.Price,
				"description": product.Description,
				"summary":     product.Summary,
				"images":      product.Images,
				"thumbnail":   product.Thumbnail,
				"category":    product.Category,
				"ingridients": product.Ingridients,
				"created_at":  product.CreatedAt,
				"updated_at":  product.UpdatedAt,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				var res map[string]interface{}
				dataBytes, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)

				err = json.Unmarshal(dataBytes, &res)
				require.NoError(t, err)

				id = res["InsertedID"].(string)
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		},
		{
			name: "duplicate a product",
			body: map[string]interface{}{
				"name":        product.Name,
				"price":       product.Price,
				"description": product.Description,
				"summary":     product.Summary,
				"images":      product.Images,
				"thumbnail":   product.Thumbnail,
				"category":    product.Category,
				"ingridients": product.Ingridients,
				"created_at":  product.CreatedAt,
				"updated_at":  product.UpdatedAt,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "create an invalid product",
			body: map[string]interface{}{},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataBytes, err := json.Marshal(test.body)
			require.NoError(t, err)

			server := NewServer(testMonogoStore)

			url := "/products"
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(dataBytes))
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}
func TestUpdateProducts(t *testing.T) {
	var tests = []struct {
		name  string
		body  map[string]interface{}
		id    string
		check func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "update a product",
			id: id,
			body: map[string]interface{}{
				"price":       product.Price,
			},
			check: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "update product by invalid id",
			id: "65bcc06cbc92379c5b6fe79b",
			body: map[string]interface{}{
				"price":       product.Price,
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

			server.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}

func TestGetAllProducts(t *testing.T) {
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

			server.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}
func TestGetProducts(t *testing.T) {
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

			server.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}

func TestDeleteProducts(t *testing.T) {
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

			server.Router.ServeHTTP(recorder, request)
			test.check(t, recorder)
		})
	}
}

func createNewProduct() Item {

	product := Item{
		Name:        "Caffe Latte",
		Price:       4.50,
		Description: "A cafe latte is a popular coffee drink that consists of espresso and steamed milk, topped with a thin layer of foam. It is perfect for those who enjoy a smooth and creamy coffee with a balanced flavor. At our coffee shop, we use high-quality beans and fresh milk to make our cafe lattes, and we can customize them with different syrups, spices, or whipped cream. ☕",
		Summary:     "A cafe latte is a coffee drink made with espresso and steamed milk, with a thin layer of foam on top. It has a smooth and creamy taste, and can be customized with different flavors. Our coffee shop offers high-quality and fresh cafe lattes for any occasion.🍵",
		Images:      []string{"caffelatte.jpeg", "lattecafe.jpeg"},
		Thumbnail:   "thumbnail.jpeg",
		Category:    "beverages",
		Ingridients: []string{"Espresso", "Milk", "Falvored syrup"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return product
}
