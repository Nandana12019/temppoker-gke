package main

import (
	"log"
	"net/http"

	"github.com/example/texas-holdem-backend/internal/api"
)

func main() {
	mux := http.NewServeMux()

	// API routes
	api.RegisterRoutes(mux)

	addr := ":8080"
	log.Printf("Starting server on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

