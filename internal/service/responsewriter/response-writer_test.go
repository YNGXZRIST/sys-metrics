package responsewriter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	WriteBadRequest(w)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("writeBadRequest error, want %v got %v", http.StatusBadRequest, res.StatusCode)
	}
}
func TestWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	WriteSuccess(w)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("writeSuccess error, want %v got %v", http.StatusOK, res.StatusCode)
	}
}

func TestWriteServerError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteServerError(w)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("writeSuccess error, want %v got %v", http.StatusInternalServerError, res.StatusCode)
	}
}

func TestWriteSuccessStatus(t *testing.T) {
	w := httptest.NewRecorder()
	WriteSuccessStatus(w)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("writeSuccessStatus error, want %v got %v", http.StatusOK, res.StatusCode)
	}
}

func TestWriteNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	WriteNotFound(w)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("WriteNotFound error, want %v got %v", http.StatusNotFound, res.StatusCode)
	}
}

func TestWriteInternalServerError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteInternalServerError(w)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("WriteInternalServerError error, want %v got %v", http.StatusInternalServerError, res.StatusCode)
	}
}
