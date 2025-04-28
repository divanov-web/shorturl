package main

import (
	"github.com/divanov-web/shorturl/internal/config"
	"github.com/divanov-web/shorturl/internal/handlers"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.NewConfig()

	h := handlers.NewHandler(cfg.BaseURL)

	r := chi.NewRouter()

	r.Post("/", h.MainPage)
	r.Get("/{id}", h.GetRealURL)

	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
