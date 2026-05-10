// Package responsewriter provides small helpers for common HTTP response status codes.
package responsewriter

import (
	"fmt"
	"net/http"
)

// WriteBadRequest responds with 400 Bad Request.
func WriteBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

// WriteSuccessStatus responds with 200 OK and no body.
func WriteSuccessStatus(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// WriteSuccess responds with 200 OK and body "200 OK".
func WriteSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("200 OK"))
	if err != nil {
		fmt.Println(err)
	}
}

// WriteServerError responds with 500 Internal Server Error.
func WriteServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}

// WriteNotFound responds with 404 Not Found.
func WriteNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

// WriteInternalServerError responds with 500 Internal Server Error (no body).
func WriteInternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}
