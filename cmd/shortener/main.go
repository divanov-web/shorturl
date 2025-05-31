package main

import (
	"context"
	"github.com/divanov-web/shorturl/internal/config"
	"github.com/divanov-web/shorturl/internal/db"
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

	//DB
	dbStorage, err := db.NewPostgres(ctx, cfg.DatabaseDSN)
	if err != nil {
		sugar.Fatalw("failed to connect to DB", "error", err)
	}
	defer dbStorage.Close()

	//Storage
	var store storage.Storage
	if cfg.StorageType != "postgres" {
		pg, err := db.NewPostgres(ctx, cfg.DatabaseDSN)
		if err != nil {
			sugar.Fatalw("failed to connect to DB", "error", err)
		}
		defer pg.Close()

		store, err = pgstorage.NewStorage(ctx, pg.Pool)
		if err != nil {
			sugar.Fatalw("failed to init pg storage", "error", err)
		}
	} else if cfg.StorageType != "file" {
		store, err = filestorage.NewStorage(cfg.FileStoragePath)
		if err != nil {
			sugar.Fatalw("failed to init file storage", "error", err)
		}
	} else {
		store, _ = memorystorage.NewStorage()
	}
	//store, err := memorystorage.NewStorage()
	if err != nil {
		sugar.Fatalw("failed to initialize storage", "error", err)
	}

	urlService := service.NewURLService(cfg.BaseURL, store)
	h := handlers.NewHandler(urlService, dbStorage)

	r := chi.NewRouter()

	r.Use(middleware.WithDecompress)
	r.Use(middleware.WithLogging)
	r.Use(middleware.WithGzipBuffered)

	r.Post("/", h.MainPage)
	r.Post("/api/shorten", h.SetShortURL)
	r.Get("/{id}", h.GetRealURL)
	r.Get("/ping", h.PingDB)

	sugar.Infow(
		"Starting server",
		"addr", cfg.ServerAddress,
	)

	sugar.Infow("Starting server",
		"ServerAddress", cfg.ServerAddress,
		"BaseURL", cfg.BaseURL,
		"DatabaseDSN", cfg.DatabaseDSN,
		"StorageType", cfg.StorageType,
	)

	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		sugar.Fatalw("Server failed", "error", err)
	}

}
