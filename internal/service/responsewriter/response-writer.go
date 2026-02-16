package responsewriter

import (
	"fmt"
	"net/http"
)

func WriteBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}
func WriteUnsupportedMediaType(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnsupportedMediaType)
}
func WriteSuccessStatus(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
func WriteSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("200 OK"))
	if err != nil {
		fmt.Println(err)
	}
}
func WriteServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}
func WriteNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}
func WriteInternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}
