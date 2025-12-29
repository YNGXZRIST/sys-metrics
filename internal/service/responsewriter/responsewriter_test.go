package responsewriter

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	WriteBadRequest(w)
	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("writeBadRequest error, want %v got %v", http.StatusBadRequest, res.StatusCode)
	}
}
func TestWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	WriteSuccess(w)
	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("writeSuccess error, want %v got %v", http.StatusOK, res.StatusCode)
	}
}

func TestWriteServerError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteServerError(w)
	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("writeSuccess error, want %v got %v", http.StatusInternalServerError, res.StatusCode)
	}
}

func TestWriteSuccessStatus(t *testing.T) {
	w := httptest.NewRecorder()
	WriteSuccessStatus(w)
	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("writeSuccessStatus error, want %v got %v", http.StatusOK, res.StatusCode)
	}
}
