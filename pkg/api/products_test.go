package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProducts(t *testing.T) {
	var tests = []struct {
		name string
		body map[string]interface{}
	}{{name: "OK", body: map[string]interface{}{
		"name":        "test",
		"price":       4.50,
		"description": "test product on a price tag of $4.50",
		"summary":     "test product",
		"images":      []string{"test1.jpeg", "test2.jpeg"},
		"thumbnail":   "test.jpeg",
		"category":    "test-category",
	}}}

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
		})
	}
}
