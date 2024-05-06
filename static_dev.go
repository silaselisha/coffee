//+build dev
//go:build dev
// +build dev

package main

import (
	"fmt"
	"net/http"
	"os"
)

func public() http.Handler {
	fmt.Println("SERVING STATIC FILES IN DEV ENVIRONMENT")
	return http.StripPrefix("/public/", http.FileServerFS(os.DirFS("public")))
}
