package main

type Store struct {
	data *coffeeList
}

func NewStore(data *coffeeList) *Store {
	return &Store{
		data: data,
	}
}
