package httpcompressor

import (
	"io"
)

// CompressorReader abstracts gzip.Reader for tests and extensions.
type CompressorReader interface {
	io.ReadCloser
	Reset(r io.Reader) error
}

// CompressReader proxies reads to the inner compressor and closes the underlying ReadCloser.
type CompressReader struct {
	r          io.ReadCloser
	compressor CompressorReader
}

func newCompressReader(r io.ReadCloser, compressor CompressorReader) (*CompressReader, error) {

	return &CompressReader{
		r:          r,
		compressor: compressor,
	}, nil
}

// Read reads decompressed data from the compressor.
func (c *CompressReader) Read(p []byte) (n int, err error) {
	return c.compressor.Read(p)
}

// Close closes the gzip layer and the underlying ReadCloser.
func (c *CompressReader) Close() error {
	if err := c.compressor.Close(); err != nil {
		return err
	}
	return c.r.Close()
}
