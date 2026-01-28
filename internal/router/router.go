package router

import (
	"sys-metrics/internal/handler"
	"sys-metrics/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func GetRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", handler.IndexHandler)
	r.Group(func(r chi.Router) {
		r.Use(middleware.ContentTypeJSON)
		r.Post("/value", handler.ValueHandlerJSON)
		r.Post("/update", handler.UpdateHandlerJSON)
	})
	r.Post("/update/{type}/{name}/{value}", handler.UpdateHandler)
	r.Get("/value/{type}/{name}", handler.ValueHandler)
	return r
}
