package httpcompressor

import (
	"io"
	"net/http"
)

// ContentEncodingHeader is the HTTP header name for the response body encoding.
const ContentEncodingHeader = "Content-Encoding"

// AcceptEncodingHeader is the HTTP header name for acceptable encodings from the client.
const AcceptEncodingHeader = "Accept-Encoding"

// Compressor abstracts gzip.Writer.
type Compressor interface {
	io.WriteCloser
	Reset(w io.Writer)
}

// CompressWriter implements http.ResponseWriter by writing gzip-compressed data and setting Content-Encoding.
type CompressWriter struct {
	w           http.ResponseWriter
	compressor  Compressor
	encoding    string
	wroteHeader bool
}

func newCompressWriter(w http.ResponseWriter, compressor Compressor, encoding string) *CompressWriter {
	compressor.Reset(w)
	return &CompressWriter{
		w:           w,
		compressor:  compressor,
		encoding:    encoding,
		wroteHeader: false,
	}
}

// Header delegates to the wrapped ResponseWriter's Header.
func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

// Write writes to gzip; on the first write it sets Content-Encoding.
func (c *CompressWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		c.w.Header().Set(ContentEncodingHeader, c.encoding)
		c.wroteHeader = true
	}
	return c.compressor.Write(p)
}

// WriteHeader sets the status code and Content-Encoding when headers have not been sent yet.
func (c *CompressWriter) WriteHeader(statusCode int) {
	if !c.wroteHeader {
		c.w.Header().Set(ContentEncodingHeader, c.encoding)
		c.wroteHeader = true
	}
	c.w.WriteHeader(statusCode)
}

// Close finishes the gzip stream.
func (c *CompressWriter) Close() error {
	return c.compressor.Close()
}
