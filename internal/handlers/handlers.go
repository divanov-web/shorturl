package handlers

import (
	"github.com/divanov-web/shorturl/internal/storage"
	"io"
	"net/http"
)

func MainPage(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/":
		handlePost(w, r)
	case r.Method == http.MethodGet && r.URL.Path != "/":
		handleGet(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	originalURL := string(body)
	shortID := storage.MakeShort(originalURL)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + shortID))
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	url, ok := storage.GetURL(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
