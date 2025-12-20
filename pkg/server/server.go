package server

import (
	"net/http"
)

const DefaultPort = "8080"
const DefaultHost = "localhost"

func NewServer(host string, port string, r *http.ServeMux) error {
	err := http.ListenAndServe(host+":"+port, r)
	return err
}
