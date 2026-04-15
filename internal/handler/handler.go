package handler

import (
	"fmt"
	"net/http"
	"strings"
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/observer"

	"go.uber.org/zap"
)

type ObserverKey string

const (
	ObserverAudit ObserverKey = "audit"
)

type Handler struct {
	Logger    *zap.Logger
	Conn      *db.DB
	Auth      authenticate.Authenticator
	Observers map[ObserverKey]observer.Observer
}

func NewHandler(c *db.DB, a authenticate.Authenticator, l *zap.Logger, observersMap map[ObserverKey]observer.Observer) *Handler {
	return &Handler{
		Logger:    l,
		Conn:      c,
		Auth:      a,
		Observers: observersMap,
	}
}
func (h *Handler) GetObserverByType(key ObserverKey) (observer.Observer, error) {
	o, ok := h.Observers[key]
	if !ok {
		return nil, fmt.Errorf("observer '%s' not found", key)
	}
	return o, nil
}
func (h *Handler) GetIPFromRequest(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}
