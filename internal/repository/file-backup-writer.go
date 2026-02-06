package repository

import (
	"bufio"
	"os"
)

type Writer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewBackupWriter(filename string) (*Writer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &Writer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}
func (w *Writer) Close() error {
	if err := w.writer.Flush(); err != nil {
		return err
	}
	return w.file.Close()
}
