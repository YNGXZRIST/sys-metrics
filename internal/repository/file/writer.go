package file

import (
	"bufio"
	"fmt"
	"os"
	"sys-metrics/internal/errors/labelerrors"
)

type Writer struct {
	file   *os.File
	writer *bufio.Writer
}

func newBackupWriter(filename string) (*Writer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to open backup file: %w", err))
	}
	return &Writer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}
func (w *Writer) Close() error {
	if err := w.writer.Flush(); err != nil {
		return labelerrors.NewLabelError("BACKUP", fmt.Errorf("failed to flush writer: %w", err))
	}
	return w.file.Close()
}
