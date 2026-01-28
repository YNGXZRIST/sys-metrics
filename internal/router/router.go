package router

import (
	"sys-metrics/internal/handler"

	"github.com/go-chi/chi/v5"
)

func GetRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", handler.IndexHandler)
	r.Post("/update/{type}/{name}/{value}", handler.UpdateHandler)
	r.Get("/value/{type}/{name}", handler.ValueHandler)
	return r
}
