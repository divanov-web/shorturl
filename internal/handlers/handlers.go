package handlers

import (
	"github.com/divanov-web/shorturl/internal/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func MainPage(w http.ResponseWriter, r *http.Request) {
	var originalURL string

	ct := r.Header.Get("Content-Type")
	switch {
	case ct == "application/x-www-form-urlencoded":
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
			return
		}
		originalURL = r.FormValue("url")
	case ct == "text/plain" || ct == "":
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "Пустое тело запроса", http.StatusBadRequest)
			return
		}
		originalURL = strings.TrimSpace(string(body))
	default:
		http.Error(w, "Неподдерживаемый Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	if originalURL == "" || !isValidURL(originalURL) {
		http.Error(w, "Некорректный URL", http.StatusBadRequest)
		return
	}

	shortID := storage.MakeShort(originalURL)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + shortID))
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func GetRealUrl(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	realUrl, ok := storage.GetURL(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, realUrl, http.StatusTemporaryRedirect)
}
