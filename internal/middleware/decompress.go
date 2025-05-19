package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// WithDecompress распаковывает запрос, если он пришёл сжатым
func WithDecompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "failed to decompress request body", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = io.NopCloser(gz) // подменяем тело запроса на распакованное
		}
		h.ServeHTTP(w, r)
	})
}
