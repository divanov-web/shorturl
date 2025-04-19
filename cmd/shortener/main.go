package main

import (
	"net/http"
)

const form = `<html>
    <head>
    <title></title>
    </head>
    <body>
        <form action="/" method="post">
            <label>Ссылка <input type="text" name="url"></label>
            <input type="submit" value="Сократить">
        </form>
    </body>
</html>`

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
		url := r.FormValue("url")
		shortUrl := makeShort(url)
		w.Write([]byte(shortUrl))
		return
	}

	if r.Method == http.MethodGet && path != "/" {
		id := path[1:]
		url := getUrl(id)
		//w.Write([]byte(url))
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	// Показываем форму
	if r.Method == http.MethodGet && path == "/" {
		w.Write([]byte(form))
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
