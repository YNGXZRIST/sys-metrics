package backup

import (
	"encoding/json"
	"sys-metrics/internal/model/metrics"
)

func (r *Reader) ReadFromBackup() ([]metrics.Metrics, error) {
	m := make([]metrics.Metrics, 0)
	for {
		line, err := r.reader.ReadString('\n')

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

func (w *Writer) WriteMetricToBackup(metric *metrics.Metrics) error {
	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	_, err = w.writer.Write(data)
	if err != nil {
		return err
	}
	return w.writer.Flush()
}
func (bc *BackupConfig) UpsertMetricToBackup(metric *metrics.Metrics) error {
	if err := bc.Reader.Reset(); err != nil {
		return err
	}

	existingMetrics, err := bc.Reader.ReadFromBackup()
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
	if err := bc.Writer.file.Truncate(0); err != nil {
		return err
	}
	if _, err := bc.Writer.file.Seek(0, 0); err != nil {
		return err
	}
	bc.Writer.writer.Reset(bc.Writer.file)
	return bc.Writer.WriteBatchMetricsToBackup(existingMetrics)
}
func (w *Writer) WriteBatchMetricsToBackup(metrics []metrics.Metrics) error {
	for _, metric := range metrics {
		data, err := json.Marshal(metric)
		if err != nil {
			return err
		}
		data = append(data, '\n')
		if _, err := w.writer.Write(data); err != nil {
			return err
		}
	}
	return w.writer.Flush()
}
func (bc *BackupConfig) UpsertBatchMetricsToBackup(newMetrics []metrics.Metrics) error {
	if err := bc.Reader.Reset(); err != nil {
		return err
	}

	existingMetrics, err := bc.Reader.ReadFromBackup()
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
	if err := bc.Writer.file.Truncate(0); err != nil {
		return err
	}
	if _, err := bc.Writer.file.Seek(0, 0); err != nil {
		return err
	}
	bc.Writer.writer.Reset(bc.Writer.file)
	return bc.Writer.WriteBatchMetricsToBackup(existingMetrics)
}
