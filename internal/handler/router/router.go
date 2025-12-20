package router

import (
	"net/http"
	"sys-metrics/internal/handler/handlers"
)

func GetRouter() *http.ServeMux {
	r := http.NewServeMux()
	r.Handle("/update/{type}/{name}/{value}", http.HandlerFunc(handlers.UpdateHandler))
	return r
}
