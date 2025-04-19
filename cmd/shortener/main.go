package main

import (
	"io"
	"net/http"
)

func makeShort(url string) string {
	url = "EwHXdJfB"
	return url
}

func getUrl(id string) string {
	return "https://practicum.yandex.ru/"
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
		shortUrl := makeShort(originalURL)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/" + shortUrl))
		return
	}

	if r.Method == http.MethodGet && path != "/" {
		id := path[1:]
		url := getUrl(id)

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
