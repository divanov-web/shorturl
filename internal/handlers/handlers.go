// Package handlers содержит HTTP-хендлеры сервиса сокращения URL
package handlers

import (
	"net/http"
	"net/url"

	"github.com/divanov-web/shorturl/internal/service"
)

// Handler набор HTTP-хендлеров сервиса.
type Handler struct {
	Service *service.URLService
}

// DBPinger пингует БД
type DBPinger interface {
	Ping() error
}

// DataRequest представляет входящие данные для создания короткой ссылки (JSON: поле "url").
type DataRequest struct {
	URL string `json:"url"`
}

// DataResponse Исходящие данные
type DataResponse struct {
	Result string `json:"result"`
}

// UserURLItem описывает пару короткий/исходный URL для ответа списка ссылок пользователя.
type UserURLItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// NewHandler создаёт Handler с переданным сервисом бизнес-логики.
func NewHandler(svc *service.URLService) *Handler {
	return &Handler{Service: svc}
}

// PingDB хэндлер. Проверяет доступность хранилища и возвращает 200 OK при успешном ответе.
func (h *Handler) PingDB(w http.ResponseWriter, r *http.Request) {
	if err := h.Service.Ping(); err != nil {
		http.Error(w, "database unavailable", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Scheme != "" && u.Host != ""
}
