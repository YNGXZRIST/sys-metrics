package file

import (
	"context"
	"encoding/json"
	"fmt"
	"sys-metrics/internal/model/metrics"
)

type BackupHandler struct {
	reader *Reader
	writer *Writer
}

func NewFileMetricsBackupHandler(reader *Reader, writer *Writer) *BackupHandler {
	return &BackupHandler{
		reader: reader,
		writer: writer,
	}
}

func (h *BackupHandler) Read(ctx context.Context) ([]metrics.Metrics, error) {
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

func (h *BackupHandler) Write(ctx context.Context, metric *metrics.Metrics) error {
	data, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("error serializing metrics: %w", err)
	}
	data = append(data, '\n')

	_, err = h.writer.writer.Write(data)
	if err != nil {
		return fmt.Errorf("error writing metrics: %w", err)
	}
	return h.writer.writer.Flush()
}

func (h *BackupHandler) WriteBatch(ctx context.Context, metrics []metrics.Metrics) error {
	for _, metric := range metrics {
		data, err := json.Marshal(metric)
		if err != nil {
			return fmt.Errorf("error serializing metrics: %w", err)
		}
		data = append(data, '\n')
		if _, err := h.writer.writer.Write(data); err != nil {
			return fmt.Errorf("error writing metrics: %w", err)
		}
	}
	return h.writer.writer.Flush()
}

func (h *BackupHandler) Upsert(ctx context.Context, metric *metrics.Metrics) error {
	if err := h.reader.Reset(); err != nil {
		return fmt.Errorf("error resetting reader: %w", err)
	}

	existingMetrics, err := h.Read(ctx)
	if err != nil {
		return fmt.Errorf("error reading metrics: %w", err)
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
		return fmt.Errorf("error truncating file: %w", err)
	}
	if _, err := h.writer.file.Seek(0, 0); err != nil {
		return fmt.Errorf("error seeking file: %w", err)
	}
	h.writer.writer.Reset(h.writer.file)

	return h.WriteBatch(ctx, existingMetrics)
}

func (h *BackupHandler) UpsertBatch(ctx context.Context, newMetrics []metrics.Metrics) error {
	if err := h.reader.Reset(); err != nil {
		return fmt.Errorf("error resetting reader: %w", err)
	}

	existingMetrics, err := h.Read(ctx)
	if err != nil {
		return fmt.Errorf("error reading metrics: %w", err)
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
		return fmt.Errorf("error truncating file: %w", err)
	}
	if _, err := h.writer.file.Seek(0, 0); err != nil {
		return fmt.Errorf("error seeking file: %w", err)
	}
	h.writer.writer.Reset(h.writer.file)

	return h.WriteBatch(ctx, existingMetrics)
}
