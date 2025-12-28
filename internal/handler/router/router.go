package router

import (
	"sys-metrics/internal/handler/handlers"

	"github.com/go-chi/chi/v5"
)

func GetRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", handlers.Index)
	r.Get("/update/{type}/{name}/{value}", handlers.UpdateHandler)
	r.Get("/value/{type}/{name}", handlers.ValueHandler)
	return r
}
