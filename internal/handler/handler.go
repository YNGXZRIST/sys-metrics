// Package handler implements the metrics server HTTP handlers (updates, reads, HTML index, ping).
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

// ObserverKey identifies an observer implementation in the Observers map.
type ObserverKey string

const (
	// ObserverAudit is the key for the metric-change audit observer.
	ObserverAudit ObserverKey = "audit"
)

// Handler holds dependencies for HTTP handlers: DB, auth, logging, and observers.
type Handler struct {
	Logger    *zap.Logger
	Conn      *db.DB
	Auth      authenticate.Authenticator
	Observers map[ObserverKey]observer.Observer
}

// NewHandler builds a Handler with optional DB connection (may be nil), authenticator, and observers map.
func NewHandler(c *db.DB, a authenticate.Authenticator, l *zap.Logger, observersMap map[ObserverKey]observer.Observer) *Handler {
	return &Handler{
		Logger:    l,
		Conn:      c,
		Auth:      a,
		Observers: observersMap,
	}
}

// GetObserverByType returns the observer for key or an error if it is not registered.
func (h *Handler) GetObserverByType(key ObserverKey) (observer.Observer, error) {
	o, ok := h.Observers[key]
	if !ok {
		return nil, fmt.Errorf("observer '%s' not found", key)
	}
	return o, nil
}

// GetIPFromRequest returns the client IP from RemoteAddr (the part before ':').
func (h *Handler) GetIPFromRequest(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}
