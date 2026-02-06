package repository

import (
	"encoding/json"
	"sys-metrics/internal/model/metrics"
)

type FileMetricsBackupHandler struct {
	reader *Reader
	writer *Writer
}

func NewFileMetricsBackupHandler(reader *Reader, writer *Writer) *FileMetricsBackupHandler {
	return &FileMetricsBackupHandler{
		reader: reader,
		writer: writer,
	}
}

func (h *FileMetricsBackupHandler) Read() ([]metrics.Metrics, error) {
	m := make([]metrics.Metrics, 0)
	for {
		line, err := h.reader.Reader.ReadString('\n')

		if len(line) > 0 {
			metric := metrics.Metrics{}
			if err := json.Unmarshal([]byte(line), &metric); err != nil {
				return nil, err
			}
			m = append(m, metric)
		}

		if err != nil {
			break
		}
	}

	return m, nil
}

func (h *FileMetricsBackupHandler) Write(metric *metrics.Metrics) error {
	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	_, err = h.writer.writer.Write(data)
	if err != nil {
		return err
	}
	return h.writer.writer.Flush()
}

func (h *FileMetricsBackupHandler) WriteBatch(metrics []metrics.Metrics) error {
	for _, metric := range metrics {
		data, err := json.Marshal(metric)
		if err != nil {
			return err
		}
		data = append(data, '\n')
		if _, err := h.writer.writer.Write(data); err != nil {
			return err
		}
	}
	return h.writer.writer.Flush()
}

func (h *FileMetricsBackupHandler) Upsert(metric *metrics.Metrics) error {
	if err := h.reader.Reset(); err != nil {
		return err
	}

	existingMetrics, err := h.Read()
	if err != nil {
		return err
	}

	found := false
	for i := range existingMetrics {
		if existingMetrics[i].ID == metric.ID {
			existingMetrics[i] = *metric
			found = true
			break
		}
	}
	if !found {
		existingMetrics = append(existingMetrics, *metric)
	}

	if err := h.writer.file.Truncate(0); err != nil {
		return err
	}
	if _, err := h.writer.file.Seek(0, 0); err != nil {
		return err
	}
	h.writer.writer.Reset(h.writer.file)

	return h.WriteBatch(existingMetrics)
}

func (h *FileMetricsBackupHandler) UpsertBatch(newMetrics []metrics.Metrics) error {
	if err := h.reader.Reset(); err != nil {
		return err
	}

	existingMetrics, err := h.Read()
	if err != nil {
		return err
	}

	metricsMap := make(map[string]int)
	for i := range existingMetrics {
		metricsMap[existingMetrics[i].ID] = i
	}

	for _, newMetric := range newMetrics {
		if idx, found := metricsMap[newMetric.ID]; found {
			existingMetrics[idx] = newMetric
		} else {
			existingMetrics = append(existingMetrics, newMetric)
			metricsMap[newMetric.ID] = len(existingMetrics) - 1
		}
	}

	if err := h.writer.file.Truncate(0); err != nil {
		return err
	}
	if _, err := h.writer.file.Seek(0, 0); err != nil {
		return err
	}
	h.writer.writer.Reset(h.writer.file)

	return h.WriteBatch(existingMetrics)
}
