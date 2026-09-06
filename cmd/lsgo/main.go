package main

import (
	"net/http"

	"github.com/mszalbach/lsgo/internal/backend"
)

func main() {
	server := http.Server{
		Addr:    ":8000",
		Handler: backend.Router(),
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
