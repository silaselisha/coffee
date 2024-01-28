package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
)

func processCoffeeData(fileName string) (results coffeeList, err error) {
	_, err = os.Stat(fileName)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	if errors.Is(err, fs.ErrNotExist) {
		log.Print(err)
		os.Exit(1)
	}

	dataInBytes, err := os.ReadFile(fileName)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	err = json.Unmarshal(dataInBytes, &results)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	return
}
