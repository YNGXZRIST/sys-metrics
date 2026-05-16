// Package router wires chi.Mux with middleware and metrics HTTP API routes.
package router

import (
	"sys-metrics/internal/handler"
	imw "sys-metrics/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// GetRouter configures gzip, request logging, pprof, authentication, and handler h routes.
func GetRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(imw.GzipMiddleware)
	r.Use(imw.WithClientIP)
	r.Use(imw.SecureMiddleware(h.Logger, h.ReqDecryptor))
	r.Use(imw.WithRequestLogger(h.Logger))
	r.Mount("/debug", chimw.Profiler())

	r.Group(func(ar chi.Router) {
		ar.Use(imw.WithAuthenticateMiddleware(h.Logger, h.Auth))

		ar.Get("/", h.IndexHandler)
		ar.Get("/ping", h.PingHandler)
		ar.Group(func(gr chi.Router) {
			gr.Use(imw.ContentTypeJSON)
			gr.Post("/value", h.ValueHandlerJSON)
			gr.Post("/value/", h.ValueHandlerJSON)
			gr.Post("/update", h.UpdateHandlerJSON)
			gr.Post("/update/", h.UpdateHandlerJSON)
			gr.Post("/updates", h.UpdatesMetricsHandlerJSON)
			gr.Post("/updates/", h.UpdatesMetricsHandlerJSON)
		})
		ar.Post("/update/{type}/{name}/{value}", h.UpdateHandler)
		ar.Get("/value/{type}/{name}", h.ValueHandler)
	})
	return r
}
