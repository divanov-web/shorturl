package handlers

import (
	"encoding/json"
	"github.com/divanov-web/shorturl/internal/service"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	Service *service.URLService
	DB      DBPinger
}

type DBPinger interface {
	Ping() error
}

// DataRequest Входящие данные
type DataRequest struct {
	URL string `json:"url"`
}

// DataResponse Исходящие данные
type DataResponse struct {
	Result string `json:"result"`
}

func NewHandler(svc *service.URLService, db DBPinger) *Handler {
	return &Handler{
		Service: svc,
		DB:      db,
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

	shortURL, err := h.Service.CreateShort(originalURL)
	if err != nil {
		http.Error(w, "Некорректный URL", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *Handler) SetShortURL(w http.ResponseWriter, r *http.Request) {
	var data DataRequest

	// Декодируем JSON напрямую из тела
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Невозможно прочитать JSON", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(data.URL)
	if originalURL == "" || !isValidURL(originalURL) {
		http.Error(w, "Некорректный URL", http.StatusBadRequest)
		return
	}

	shortURL, err := h.Service.CreateShort(originalURL)
	if err != nil {
		http.Error(w, "Некорректный URL", http.StatusBadRequest)
		return
	}
	result := DataResponse{Result: shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
	}
}

// GetRealURL хэндлер Get запрос на получение ссылки из хеша
func (h *Handler) GetRealURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	realURL, ok := h.Service.ResolveShort(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, realURL, http.StatusTemporaryRedirect)
}

// PingDB хэндлер пинга ДБ
func (h *Handler) PingDB(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.Ping(); err != nil {
		http.Error(w, "database unavailable", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Scheme != "" && u.Host != ""
}
