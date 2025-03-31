//go:build !netlify

package main

import (
	"log"
	"net/http"
)

func start(handler http.Handler) {
	log.Printf("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
