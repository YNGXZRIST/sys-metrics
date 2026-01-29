package router

import (
	"sys-metrics/internal/handler"
	"sys-metrics/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func GetRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", handler.IndexHandler)
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
