package main

import (
	"log"
	"net/http"
	"os"
)

const SERVER_ADDRES = ":3000"

func main() {
	list, err := processCoffeeData("coffee.json")
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	s := NewStore(&list)

  http.HandleFunc("/coffee", s.getAllCoffeeHandler)
	err = http.ListenAndServe(SERVER_ADDRES, nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
