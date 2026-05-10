// Package httpcompressor wraps HTTP bodies with gzip-compressing writers and decompressing readers.
package httpcompressor

import (
	"compress/gzip"
	"io"
	"net/http"
)

// GzipEncoding is the Content-Encoding / Accept-Encoding token for gzip.
const GzipEncoding = "gzip"

// NewGzipReader returns a reader that decompresses gzip data from r.
func NewGzipReader(r io.ReadCloser) (*CompressReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return newCompressReader(r, gr)
}

// NewGzipWriter returns a ResponseWriter that gzip-compresses the response body.
func NewGzipWriter(w http.ResponseWriter) *CompressWriter {
	gw := gzip.NewWriter(w)
	return newCompressWriter(w, gw, GzipEncoding)
}
