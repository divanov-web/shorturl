package main

import (
	"github.com/divanov-web/shorturl/internal/config"
	"github.com/divanov-web/shorturl/internal/handlers"
	"github.com/divanov-web/shorturl/internal/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()

	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	// делаем регистратор SugaredLogger
	sugar := logger.Sugar()
	middleware.SetLogger(sugar) // передаём логгер в middleware

	h := handlers.NewHandler(cfg.BaseURL)

	r := chi.NewRouter()

	r.Use(middleware.WithLogging)

	r.Post("/", h.MainPage)
	r.Get("/{id}", h.GetRealURL)

	sugar.Infow(
		"Starting server",
		"addr", cfg.ServerAddress,
	)

	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		sugar.Fatalw("Server failed", "error", err)
	}

}
