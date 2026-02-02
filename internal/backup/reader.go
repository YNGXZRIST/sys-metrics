package backup

import (
	"bufio"
	"os"
)

type Reader struct {
	file   *os.File
	reader *bufio.Reader
}

func newBackupReader(filename string) (*Reader, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	return &Reader{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}
func (r *Reader) Reset() error {
	_, err := r.file.Seek(0, 0)
	if err != nil {
		return err
	}
	r.reader.Reset(r.file)
	return nil
}
func (r *Reader) Close() error {
	return r.file.Close()
}
