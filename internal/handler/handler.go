// Package handler implements the metrics server HTTP handlers (updates, reads, HTML index, ping).
package handler

import (
	"fmt"
	"sync"
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/observer"
	"sys-metrics/internal/secure"
	"sys-metrics/internal/service/metrics"

	"go.uber.org/zap"
)

// ObserverKey identifies an observer implementation in the Observers map.
type ObserverKey string

const (
	// ObserverAudit is the key for the metric-change audit observer.
	ObserverAudit ObserverKey = "audit"
)

// generate:reset

// Handler holds dependencies for HTTP handlers: DB, auth, logging, and observers.
type Handler struct {
	Logger        *zap.Logger
	Conn          *db.DB
	Auth          authenticate.Authenticator
	ReqDecryptor  *secure.RequestDecryptor
	Observers     map[ObserverKey]observer.Observer
	mu            sync.Mutex
	MetricService *metrics.MetricService
}

// NewHandler builds a Handler with optional DB connection (maybe nil), authenticator, observers map, and metrics service.
func NewHandler(c *db.DB, a authenticate.Authenticator, d *secure.RequestDecryptor, l *zap.Logger, observersMap map[ObserverKey]observer.Observer, ms *metrics.MetricService) *Handler {
	return &Handler{
		Logger:        l,
		Conn:          c,
		Auth:          a,
		ReqDecryptor:  d,
		Observers:     observersMap,
		MetricService: ms,
		mu:            sync.Mutex{},
	}
}

// GetObserverByType returns the observer for key or an error if it is not registered.
func (h *Handler) GetObserverByType(key ObserverKey) (observer.Observer, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	o, ok := h.Observers[key]
	if !ok {
		return nil, fmt.Errorf("observer '%s' not found", key)
	}
	return o, nil
}
