package postgress

import (
	"context"
	"database/sql"
	"fmt"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/model/metrics"
)

type Handler struct {
	dbConn *db.DB
}

func NewHandler(dbConn *db.DB) *Handler {
	return &Handler{dbConn: dbConn}
}

func (h Handler) Upsert(ctx context.Context, metric *metrics.Metrics) error {
	var updateColumn string
	switch metric.MType {
	case common.Gauge:
		updateColumn = "value"
	case common.Counter:
		updateColumn = "delta"
	default:
		return fmt.Errorf("unsupported metric type: %s", metric.MType)
	}
	sqlStatement := fmt.Sprintf(`
		INSERT INTO metrics (id, mtype, delta, value, hash)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id)
		DO UPDATE SET %s = EXCLUDED.%s`,
		updateColumn, updateColumn)
	_, err := h.dbConn.ExecContext(ctx, sqlStatement, metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash)
	if err != nil {
		return fmt.Errorf("upsert failed: %w", err)
	}
	return nil
}

func (h Handler) Read(ctx context.Context) ([]metrics.Metrics, error) {
	rows, err := h.dbConn.QueryContext(ctx, "SELECT id, mtype, delta, value, hash FROM metrics")
	if err != nil {
		return nil, fmt.Errorf("query metrics: %w", err)
	}
	defer rows.Close()
	var result []metrics.Metrics
	for rows.Next() {
		var m metrics.Metrics
		var hash sql.NullString
		err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &hash)
		if err != nil {
			return nil, fmt.Errorf("scan metrics: %w", err)
		}
		if hash.Valid {
			m.Hash = hash.String
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (h Handler) Write(ctx context.Context, metric *metrics.Metrics) error {
	err := h.Upsert(ctx, metric)
	if err != nil {
		return fmt.Errorf("upsert metrics: %w", err)
	}
	return nil
}

func (h Handler) WriteBatch(ctx context.Context, metrics []metrics.Metrics) error {
	for _, metric := range metrics {
		err := h.Upsert(ctx, &metric)
		if err != nil {
			return fmt.Errorf("upsert metrics: %w", err)
		}
	}
	return nil
}
