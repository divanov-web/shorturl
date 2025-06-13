package main

import (
	"context"
	"github.com/divanov-web/shorturl/internal/config"
	"github.com/divanov-web/shorturl/internal/handlers"
	"github.com/divanov-web/shorturl/internal/middleware"
	"github.com/divanov-web/shorturl/internal/service"
	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/divanov-web/shorturl/internal/storage/filestorage"
	"github.com/divanov-web/shorturl/internal/storage/memorystorage"
	"github.com/divanov-web/shorturl/internal/storage/pgstorage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	cfg := config.NewConfig()

	// создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// делаем регистратор SugaredLogger
	sugar := logger.Sugar()
	middleware.SetLogger(sugar) // передаём логгер в middleware
	//сброс буфера логгера (добавлено про запас по урокам)
	defer func() {
		if err := logger.Sync(); err != nil {
			sugar.Errorw("Failed to sync logger", "error", err)
		}
	}()

	//context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := initStorage(ctx, cfg)
	if err != nil {
		sugar.Fatalw("failed to initialize storage", "error", err)
	}

	urlService := service.NewURLService(cfg.BaseURL, store)
	h := handlers.NewHandler(urlService)

	r := chi.NewRouter()

	r.Use(middleware.WithDecompress)
	r.Use(middleware.WithLogging)      //логирование
	r.Use(middleware.WithGzipBuffered) //сжатие

	auth := middleware.NewAuth(cfg.AuthSecret) //авторизация
	r.Use(auth.WithAuth)

	r.Post("/", h.MainPage)
	r.Post("/api/shorten", h.SetShortURL)
	r.Get("/{id}", h.GetRealURL)
	r.Get("/ping", h.PingDB)
	r.Post("/api/shorten/batch", h.SetShortenBatch)

	sugar.Infow(
		"Starting server",
		"addr", cfg.ServerAddress,
	)

	sugar.Infow("Config",
		"ServerAddress", cfg.ServerAddress,
		"BaseURL", cfg.BaseURL,
		"DatabaseDSN", cfg.DatabaseDSN,
		"FileStoragePath", cfg.FileStoragePath,
		"StorageType", cfg.StorageType,
	)

	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		sugar.Fatalw("Server failed", "error", err)
	}

}

func initStorage(ctx context.Context, cfg *config.Config) (storage.Storage, error) {
	switch cfg.StorageType {
	case "postgres":
		pool, err := pgstorage.NewPool(ctx, cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}
		store, err := pgstorage.NewStorage(ctx, pool)
		if err != nil {
			pool.Close()
			return nil, err
		}
		return store, nil

	case "file":
		return filestorage.NewStorage(cfg.FileStoragePath)

	default:
		return memorystorage.NewStorage()
	}
}
