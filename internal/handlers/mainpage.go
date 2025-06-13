package handlers

import (
	"errors"
	"github.com/divanov-web/shorturl/internal/middleware"
	"github.com/divanov-web/shorturl/internal/service"
	"io"
	"net/http"
	"strings"
)

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

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Ошибка определения userID из кук", http.StatusBadRequest)
		return
	}

	shortURL, err := h.Service.CreateShort(userID, originalURL)
	if errors.Is(err, service.ErrAlreadyExists) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusConflict) // 409
		w.Write([]byte(shortURL))
		return
	}
	if err != nil {
		http.Error(w, "Некорректный URL", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}
