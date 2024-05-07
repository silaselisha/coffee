//go:build dev
// +build dev

package main

import (
	"fmt"
	"net/http"
	"os"
)

func public() http.Handler {
	fmt.Println("DEV MODE: **FILE SERVER**")
	return http.StripPrefix("/public/", http.FileServerFS(os.DirFS("public")))
}
