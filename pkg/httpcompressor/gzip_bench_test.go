package httpcompressor

import (
	"net/http/httptest"
	"testing"
)

func BenchmarkGzipWriter_WriteSmall(b *testing.B) {
	payload := []byte(`{"id":"x","type":"gauge","value":123.45}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		cw := NewGzipWriter(rec)
		_, _ = cw.Write(payload)
		_ = cw.Close()
	}
}

func BenchmarkGzipWriter_WriteMedium(b *testing.B) {
	payload := make([]byte, 4096)
	for i := range payload {
		payload[i] = byte('a' + i%26)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		cw := NewGzipWriter(rec)
		_, _ = cw.Write(payload)
		_ = cw.Close()
	}
}
