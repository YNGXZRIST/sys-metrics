package httpcompressor

import (
	"io"
)

type Compressor interface {
	io.WriteCloser
	Reset(w io.Writer)
}
