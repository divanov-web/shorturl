package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type bufferedResponseWriter struct {
	http.ResponseWriter
	status      int
	header      http.Header
	body        *strings.Builder
	wroteHeader bool
}

func newBufferedResponseWriter(w http.ResponseWriter) *bufferedResponseWriter {
	return &bufferedResponseWriter{
		ResponseWriter: w,
		header:         make(http.Header),
		body:           &strings.Builder{},
	}
}

// Header возвращает буферизованные заголовки ответа
func (b *bufferedResponseWriter) Header() http.Header {
	return b.header
}

// WriteHeader сохраняет код статуса в буфере, но не отправляет его сразу клиенту.
func (b *bufferedResponseWriter) WriteHeader(statusCode int) {
	if !b.wroteHeader {
		b.status = statusCode
		b.wroteHeader = true
	}
}

// Write записывает данные в буфер тела ответа.
func (b *bufferedResponseWriter) Write(data []byte) (int, error) {
	return b.body.Write(data)
}

// WithGzipBuffered сжимаем если Accept-Encoding: gzip и response Content-Type = JSON or HTML
func WithGzipBuffered(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		// буферизуем ответ
		bufWriter := newBufferedResponseWriter(w)
		h.ServeHTTP(bufWriter, r)

		contentType := bufWriter.header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/html") {
			// включаем gzip
			w.Header().Set("Content-Encoding", "gzip")
			// копируем заголовки из буфера
			for k, v := range bufWriter.header {
				w.Header()[k] = v
			}
			w.WriteHeader(bufWriter.status)

			gz := gzip.NewWriter(w)
			defer gz.Close()
			io.WriteString(gz, bufWriter.body.String())
		} else {
			// отдаём как есть, без gzip
			for k, v := range bufWriter.header {
				w.Header()[k] = v
			}
			w.WriteHeader(bufWriter.status)
			io.WriteString(w, bufWriter.body.String())
		}
	})
}
