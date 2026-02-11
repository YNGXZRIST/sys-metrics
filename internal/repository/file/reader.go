package file

import (
	"bufio"
	"fmt"
	"os"
)

type Reader struct {
	file   *os.File
	Reader *bufio.Reader
}

func NewBackupReader(filename string) (*Reader, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	return &Reader{
		file:   file,
		Reader: bufio.NewReader(file),
	}, nil
}
func (r *Reader) Reset() error {
	_, err := r.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error reset file backup: %w", err)
	}
	r.Reader.Reset(r.file)
	return nil
}
func (r *Reader) Close() error {
	return r.file.Close()
}
