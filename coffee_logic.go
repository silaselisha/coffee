package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func (s *Store) getAllCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	data := s.data
	encoder := json.NewEncoder(w)
	err := encoder.Encode(data)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
