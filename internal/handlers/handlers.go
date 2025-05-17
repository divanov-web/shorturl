package handlers

import (
	"encoding/json"
	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	BaseURL string
}

// DataRequest Входящие данные
type DataRequest struct {
	Url string `json:"url"`
}

// DataResponse Исходящие данные
type DataResponse struct {
	Result string `json:"result"`
}

func NewHandler(baseURL string) *Handler {
	return &Handler{
		BaseURL: baseURL,
	}
}

// MainPage POST запрос на отправку большой ссылки и возвращение короткой ссылки в виде хеша
func (h *Handler) MainPage(w http.ResponseWriter, r *http.Request) {
	var originalURL string

	ct := r.Header.Get("Content-Type")
	switch {
	case ct == "application/x-www-form-urlencoded":
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
			return
		}
		originalURL = r.FormValue("url")
	default:
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "Пустое тело запроса", http.StatusBadRequest)
			return
		}
		originalURL = strings.TrimSpace(string(body))
	}

	if originalURL == "" || !isValidURL(originalURL) {
		http.Error(w, "Некорректный URL", http.StatusBadRequest)
		return
	}

	shortID := storage.MakeShort(originalURL)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.BaseURL + "/" + shortID))
}

func (h *Handler) SetShortURL(w http.ResponseWriter, r *http.Request) {
	var data DataRequest

	// Декодируем JSON напрямую из тела
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Невозможно прочитать JSON", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(data.Url)
	if originalURL == "" || !isValidURL(originalURL) {
		http.Error(w, "Некорректный URL", http.StatusBadRequest)
		return
	}

	shortID := storage.MakeShort(originalURL)
	result := DataResponse{
		Result: h.BaseURL + "/" + shortID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
	}
}

// GetRealURL Get запрос на получение ссылки из хеша
func (h *Handler) GetRealURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	realURL, ok := storage.GetURL(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, realURL, http.StatusTemporaryRedirect)
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Scheme != "" && u.Host != ""
}
