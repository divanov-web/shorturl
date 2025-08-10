package handlers_test

import (
	"bytes"
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/divanov-web/shorturl/internal/handlers"
	"github.com/divanov-web/shorturl/internal/middleware"
	"github.com/divanov-web/shorturl/internal/service"
	"github.com/divanov-web/shorturl/internal/storage/memorystorage"
)

// ExamplePingDB показывает вызов эндпоинта /ping.
func ExampleHandler_PingDB() {
	store := memorystorage.NewTestStorage()
	svc := service.NewURLService(context.Background(), "http://localhost:8080", store)
	h := handlers.NewHandler(svc)

	// Регистрируем только /ping
	r := chi.NewRouter()
	r.Get("/ping", h.PingDB)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Печатаем код ответа
	// Output: 200
	println(rr.Code)
}

// Example_createAndRedirect показывает создание короткой ссылки через POST "/"
// и переход по ней через GET "/{id}".
func Example_createAndRedirect() {
	// Сервис и хэндлер
	store := memorystorage.NewTestStorage()
	svc := service.NewURLService(context.Background(), "http://localhost:8080", store)
	h := handlers.NewHandler(svc)

	// Маршруты с авторизацией: NewAuth создаст userID в cookie автоматически
	auth := middleware.NewAuth("dev-secret-key")

	r := chi.NewRouter()
	r.Use(auth.WithAuth)
	r.Post("/", h.MainPage)
	r.Get("/{id}", h.GetRealURL)

	// Создаём короткую ссылку (text/plain)
	postReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://example.com"))
	postReq.Header.Set("Content-Type", "text/plain")
	postRec := httptest.NewRecorder()
	r.ServeHTTP(postRec, postReq)

	// Печатаем статус создания (ожидается 201)
	println(postRec.Code)

	// Извлекаем id из ответа: в теле лежит полный короткий URL
	shortURL := strings.TrimSpace(postRec.Body.String())
	id := strings.TrimPrefix(shortURL, "http://localhost:8080/")
	if id == shortURL { // на случай другой базовой ссылки
		parts := strings.Split(shortURL, "/")
		id = parts[len(parts)-1]
	}

	//Запрашиваем редирект по короткой ссылке
	getReq := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)

	// Печатаем статус редиректа (обычно 307)
	// Output:
	// 201
	// 307
	println(getRec.Code)
}
