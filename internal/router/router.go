package router

import (
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/handler"
	"sys-metrics/internal/middleware"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func GetRouter(logger *zap.Logger, conn *db.DB) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.WithDBContext(conn))
	r.Use(middleware.WithRequestLogger(logger))
	r.Use(middleware.WithLoggerContext(logger))
	r.Use(middleware.WithDBContext(conn))
	r.Get("/", handler.IndexHandler)
	r.Get("/ping", handler.PingHandler)
	r.Group(func(gr chi.Router) {
		gr.Use(middleware.ContentTypeJSON)
		gr.Post("/value", handler.ValueHandlerJSON)
		gr.Post("/update", handler.UpdateHandlerJSON)
		gr.Post("/value/", handler.ValueHandlerJSON)
		gr.Post("/update/", handler.UpdateHandlerJSON)
	})
	r.Post("/update/{type}/{name}/{value}", handler.UpdateHandler)
	r.Get("/value/{type}/{name}", handler.ValueHandler)
	return r
}
