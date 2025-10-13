package main

import (
	"log"
	"net/http"
	"shorted/internal/repository/memory"
	"shorted/internal/service/shortener"
	initRouters "shorted/internal/transport/http"
)

func main() {
	linkRepo := memory.NewLinkRepo()
	shortenerService := shortener.NewService(linkRepo)
	router := initRouters.NewRouter(shortenerService)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Println("Server starting on :8080")
	log.Fatal(server.ListenAndServe())
}
