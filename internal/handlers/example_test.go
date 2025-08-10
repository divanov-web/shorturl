package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
func ExamplePingDB() {
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

// ExampleHandler_SetShortURL демонстрирует создание короткой ссылки через POST /api/shorten.
func ExampleHandler_SetShortURL() {

	svc := service.NewURLService(nil, "http://localhost:8080", memorystorage.NewTestStorage())
	h := handlers.NewHandler(svc)

	body := bytes.NewBufferString(`{"url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	w := httptest.NewRecorder()

	h.SetShortURL(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())

	// Output:
	// 201
	// {"result":"http://localhost:8080/<shortid>"}
}

// ExampleHandler_GetRealURL демонстрирует получение оригинального URL по короткому идентификатору.
func ExampleHandler_GetRealURL() {
	svc := service.NewURLService(nil, "http://localhost:8080", memorystorage.NewTestStorage())
	h := handlers.NewHandler(svc)

	// Добавляем тестовую запись
	id, _ := svc.CreateShort("user1", "https://example.com")

	req := httptest.NewRequest(http.MethodGet, "/"+id, nil)
	w := httptest.NewRecorder()

	h.GetRealURL(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Header().Get("Location"))

	// Output:
	// 307
	// https://example.com
}

// ExampleHandler_SetShortenBatch демонстрирует сокращение списка ссылок через POST /api/shorten/batch.
func ExampleHandler_SetShortenBatch() {
	svc := service.NewURLService(nil, "http://localhost:8080", memorystorage.NewTestStorage())
	h := handlers.NewHandler(svc)

	items := []service.BatchRequestItem{
		{CorrelationID: "1", OriginalURL: "https://a.com"},
		{CorrelationID: "2", OriginalURL: "https://b.com"},
	}
	body, _ := json.Marshal(items)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.SetShortenBatch(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Body.Len() > 0)

	// Output:
	// 201
	// true
}

// ExampleHandler_GetUserURLs демонстрирует получение всех ссылок пользователя.
func ExampleHandler_GetUserURLs() {
	svc := service.NewURLService(nil, "http://localhost:8080", memorystorage.NewTestStorage())
	h := handlers.NewHandler(svc)

	// Заполним тестовыми данными
	svc.CreateShort("user1", "https://example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	w := httptest.NewRecorder()

	h.GetUserURLs(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Body.Len() > 0)

	// Output:
	// 200
	// true
}

// ExampleHandler_DeleteUserURL демонстрирует удаление ссылок пользователя.
func ExampleHandler_DeleteUserURL() {
	svc := service.NewURLService(nil, "http://localhost:8080", memorystorage.NewTestStorage())
	h := handlers.NewHandler(svc)

	// Добавляем тестовые данные
	id, _ := svc.CreateShort("user1", "https://example.com")
	body, _ := json.Marshal([]string{id})

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.DeleteUserURL(w, req)

	fmt.Println(w.Code)

	// Output:
	// 202
}

// TestExamplesPlaceholder
//func TestExamplesPlaceholder(t *testing.T) {}
