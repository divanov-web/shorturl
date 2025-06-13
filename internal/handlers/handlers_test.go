package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/divanov-web/shorturl/internal/config"
	"github.com/divanov-web/shorturl/internal/middleware"
	"github.com/divanov-web/shorturl/internal/service"
	"github.com/divanov-web/shorturl/internal/storage/filestorage"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAuthSecret = "my-secret-key"

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
			store := filestorage.NewTestStorage()
			svc := service.NewURLService("http://localhost:8080", store)
			h := NewHandler(svc)

			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "test-user-id")) // добавили userID
			w := httptest.NewRecorder()
			h.MainPage(w, req)

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
		setup  func(*filestorage.Storage)
		want   want
	}{
		{
			name:   "existing ID",
			method: http.MethodGet,
			path:   "/abc123",
			setup: func(s *filestorage.Storage) {
				s.ForceSet("abc123", "https://example.com")
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
			setup:  func(s *filestorage.Storage) {},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := filestorage.NewTestStorage()
			tt.setup(store)
			svc := service.NewURLService("http://localhost:8080", store)
			h := NewHandler(svc)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Get("/{id}", h.GetRealURL)
			r.ServeHTTP(w, req)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, result.Header.Get("Location"))
			}
		})
	}
}

func TestSetShortURL(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()
	middleware.SetLogger(sugar)

	cfg := config.NewConfig()
	store := filestorage.NewTestStorage()
	svc := service.NewURLService(cfg.BaseURL, store)
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Use(middleware.WithLogging)
	auth := middleware.NewAuth(cfg.AuthSecret) //авторизация
	r.Use(auth.WithAuth)
	r.Post("/api/shorten", h.SetShortURL)

	requestBody, err := json.Marshal(map[string]string{"url": "https://practicum.yandex.ru"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Добавляем auth_token
	token := generateTestJWT("test-user-id")
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	contentType := rec.Header().Get("Content-Type")
	assert.Contains(t, contentType, "application/json")

	var resp DataResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Result)
	assert.True(t, strings.HasPrefix(resp.Result, cfg.BaseURL+"/"))
}

func generateTestJWT(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	signed, _ := token.SignedString([]byte(testAuthSecret))
	return signed
}
