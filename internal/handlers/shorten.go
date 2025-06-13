package handlers

import (
	"encoding/json"
	"errors"
	"github.com/divanov-web/shorturl/internal/middleware"
	"github.com/divanov-web/shorturl/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

func (h *Handler) SetShortURL(w http.ResponseWriter, r *http.Request) {
	var data DataRequest

	// Декодируем JSON напрямую из тела
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Невозможно прочитать JSON", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(data.URL)
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
		result := DataResponse{Result: shortURL}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict) // 409
		_ = json.NewEncoder(w).Encode(result)
		return
	}
	if err != nil {
		http.Error(w, "Ошибка преобразования URL", http.StatusBadRequest)
		return
	}
	result := DataResponse{Result: shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
	}
}

// GetRealURL хэндлер Get запрос на получение ссылки из хеша
func (h *Handler) GetRealURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	realURL, ok := h.Service.ResolveShort(id)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, realURL, http.StatusTemporaryRedirect)
}

// SetShortenBatch обрабатывает POST /api/shorten/batch
func (h *Handler) SetShortenBatch(w http.ResponseWriter, r *http.Request) {
	var batch []service.BatchRequestItem

	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "Невозможно прочитать JSON", http.StatusBadRequest)
		return
	}

	if len(batch) == 0 {
		http.Error(w, "Пустой список", http.StatusBadRequest)
		return
	}

	// Валидация URL
	for _, item := range batch {
		if !isValidURL(item.OriginalURL) {
			http.Error(w, "Некорректный URL", http.StatusBadRequest)
			return
		}
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Ошибка определения userID из кук", http.StatusBadRequest)
		return
	}

	results, err := h.Service.CreateShortBatch(userID, batch)
	if err != nil {
		http.Error(w, "Ошибка при сохранении ссылок", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Ошибка сериализации ответа", http.StatusInternalServerError)
	}
}

// GetUserURLs хэндлер Get запрос на получение списка url текущего юзера
func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	urls, err := h.Service.GetUserURLs(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := make([]UserURLItem, 0, len(urls))
	for _, item := range urls {
		response = append(response, UserURLItem{
			ShortURL:    h.Service.BaseURL + "/" + item.ShortURL,
			OriginalURL: item.OriginalURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
