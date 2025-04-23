package main

import (
	"github.com/divanov-web/shorturl/internal/handlers"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Post("/", handlers.MainPage) // POST /car
	r.Get("/{id}", handlers.GetRealUrl)

	log.Fatal(http.ListenAndServe(":8080", r))
}
