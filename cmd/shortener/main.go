package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof

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
)

// билдить командой
// go build -o shortener.exe -ldflags "-X 'main.buildVersion=v1.0.0' -X 'main.buildDate=2025-08-19' -X 'main.buildCommit=iter20'" ./cmd/shortener
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Printf("Build version: %s\n", valueOrNA(buildVersion))
	fmt.Printf("Build date: %s\n", valueOrNA(buildDate))
	fmt.Printf("Build commit: %s\n", valueOrNA(buildCommit))

	cfg := config.NewConfig()

	//Сервер профилирования pprof
	if cfg.PprofMode {
		go func() {
			log.Println("pprof enabled at http://localhost:6060/debug/pprof/")
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				log.Printf("pprof server error: %v", err)
			}
		}()
	}

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
		if syncErr := logger.Sync(); syncErr != nil {
			sugar.Errorw("Failed to sync logger", "error", syncErr)
		}
	}()

	//context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := initStorage(ctx, cfg)
	if err != nil {
		sugar.Fatalw("failed to initialize storage", "error", err)
	}

	urlService := service.NewURLService(ctx, cfg.BaseURL, store)
	h := handlers.NewHandler(urlService)

	r := chi.NewRouter()

	r.Use(middleware.WithDecompress)
	r.Use(middleware.WithLogging)      //логирование
	r.Use(middleware.WithGzipBuffered) //сжатие

	auth := middleware.NewAuth(cfg.AuthSecret) //авторизация
	r.Use(auth.WithAuth)

	r.Post("/", h.MainPage)                         //Сохранение url с request текстовых параметров
	r.Post("/api/shorten", h.SetShortURL)           //Сохранение url с request json параметров
	r.Get("/{id}", h.GetRealURL)                    //Вернуть исходных url по его хешу и сделать редирект
	r.Get("/ping", h.PingDB)                        // пингует БД постгресс
	r.Post("/api/shorten/batch", h.SetShortenBatch) //Сохранение пачки url
	r.Get("/api/user/urls", h.GetUserURLs)          //Получить все url пользователя
	r.Delete("/api/user/urls", h.DeleteUserURL)     //Удалить url пользователя по массиву id

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
		"PprofMode", cfg.PprofMode,
		"EnableHTTPS", cfg.EnableHTTPS,
	)

	if cfg.EnableHTTPS {
		sugar.Infow("HTTPS enabled, starting with TLS")
		if err := http.ListenAndServeTLS(cfg.ServerAddress, "server.crt", "server.key", r); err != nil {
			sugar.Fatalw("Server failed", "error", err)
		}
	} else {
		if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
			sugar.Fatalw("Server failed", "error", err)
		}
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

func valueOrNA(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}
