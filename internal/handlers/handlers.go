package handlers

import (
	"github.com/divanov-web/shorturl/internal/service"
	"net/http"
	"net/url"
)

type Handler struct {
	Service *service.URLService
}

type DBPinger interface {
	Ping() error
}

// DataRequest Входящие данные
type DataRequest struct {
	URL string `json:"url"`
}

// DataResponse Исходящие данные
type DataResponse struct {
	Result string `json:"result"`
}

func NewHandler(svc *service.URLService) *Handler {
	return &Handler{Service: svc}
}

// PingDB хэндлер пинга ДБ
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
