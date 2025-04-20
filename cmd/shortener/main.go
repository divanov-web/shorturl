package main

import (
	"io"
	"math/rand"
	"net/http"
	"time"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
var urlStorage = make(map[string]string)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rnd.Intn(len(charset))]
	}
	return string(b)
}

func makeShort(url string) string {
	id := generateID(8)
	urlStorage[id] = url
	return id
}

func getURL(id string) (string, bool) {
	url, ok := urlStorage[id]
	return url, ok
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if r.Method == http.MethodPost && path == "/" {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		originalURL := string(body)
		shortID := makeShort(originalURL)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/" + shortID))
		return
	}

	if r.Method == http.MethodGet && path != "/" {
		id := path[1:]
		url, ok := getURL(id)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainPage)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
