package handlers

import (
	"github.com/divanov-web/shorturl/internal/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlePost(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		bodyPrefix  string
	}

	tests := []struct {
		name   string
		method string
		body   string
		want   want
	}{
		{
			name:   "valid URL",
			method: http.MethodPost,
			body:   "https://example.com",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
				bodyPrefix:  "http://localhost:8080/",
			},
		},
		{
			name:   "empty body",
			method: http.MethodPost,
			body:   "",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handlePost(w, req)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			}

			if tt.want.bodyPrefix != "" {
				body, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(string(body), tt.want.bodyPrefix), "unexpected response body: %s", body)
			}
		})
	}
}

func TestHandleGet(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}

	tests := []struct {
		name   string
		method string
		path   string
		setup  func()
		want   want
	}{
		{
			name:   "existing ID",
			method: http.MethodGet,
			path:   "/abc123",
			setup: func() {
				storage.ForceSet("abc123", "https://example.com")
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com",
			},
		},
		{
			name:   "non-existent ID",
			method: http.MethodGet,
			path:   "/doesnotexist",
			setup:  func() {}, // ничего не делаем
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handleGet(w, req)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, result.Header.Get("Location"))
			}
		})
	}
}
