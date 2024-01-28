package main

type coffee struct {
	Id          _id    `json:"_id"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	Weight      int32  `json:"weight"`
	Manufacture string `json:"manufacture"`
	Grade       int32  `json:"grade"`
	ExpiryDate  string `json:"expiry_date"`
}

type _id int64
type coffeeList []*coffee
