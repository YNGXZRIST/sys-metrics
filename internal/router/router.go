package router

import (
	"sys-metrics/internal/handler"
	"sys-metrics/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func GetRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.WithRequestLogger(h.Logger))
	r.Use(middleware.WithAuthenticateMiddleware(h.Logger, h.Auth))
	r.Get("/", h.IndexHandler)
	r.Get("/ping", h.PingHandler)
	r.Group(func(gr chi.Router) {
		gr.Use(middleware.ContentTypeJSON)
		gr.Post("/value", h.ValueHandlerJSON)
		gr.Post("/value/", h.ValueHandlerJSON)
		gr.Post("/update", h.UpdateHandlerJSON)
		gr.Post("/update/", h.UpdateHandlerJSON)
		gr.Post("/updates", h.UpdatesMetricsHandlerJSON)
		gr.Post("/updates/", h.UpdatesMetricsHandlerJSON)
	})
	r.Post("/update/{type}/{name}/{value}", h.UpdateHandler)
	r.Get("/value/{type}/{name}", h.ValueHandler)
	return r
}
