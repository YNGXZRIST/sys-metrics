package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
	"sys-metrics/internal/model/metrics"
)

const limit = 100
const batchSize = 100

type Handler struct {
	dbConn *db.DB
}

func NewHandler(dbConn *db.DB) *Handler {
	return &Handler{dbConn: dbConn}
}

func (h *Handler) Upsert(ctx context.Context, metric *metrics.Metrics) error {
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
		INSERT INTO metrics (name, mtype, delta, value, hash)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (name)
		DO UPDATE SET %s = EXCLUDED.%s`,
		updateColumn, updateColumn)
	_, err := h.dbConn.ExecContext(ctx, sqlStatement, metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash)
	if err != nil {
		return fmt.Errorf("upsert failed: %w", err)
	}
	return nil
}

func (h *Handler) Read(ctx context.Context) ([]metrics.Metrics, error) {

	var result []metrics.Metrics
	lastID := int64(0)
	for {
		batch, nextID, err := h.chunkSelect(ctx, lastID)
		if err != nil {
			return nil, err
		}
		result = append(result, batch...)
		if len(batch) < limit {
			break
		}
		lastID = nextID
	}
	return result, nil
}

func (h *Handler) chunkSelect(ctx context.Context, lastID int64) ([]metrics.Metrics, int64, error) {
	rows, err := h.dbConn.QueryContext(ctx,
		"SELECT id, name, mtype, delta, value, hash FROM metrics WHERE id > $1 ORDER BY id LIMIT $2",
		lastID, limit)
	if err != nil {
		return nil, lastID, fmt.Errorf("query metrics: %w", err)
	}
	defer rows.Close()
	var batch []metrics.Metrics
	var nextID = lastID
	for rows.Next() {
		var m metrics.Metrics
		var hash sql.NullString
		var rowID int64
		err = rows.Scan(&rowID, &m.ID, &m.MType, &m.Delta, &m.Value, &hash)
		if err != nil {
			return nil, lastID, fmt.Errorf("scan metrics: %w", err)
		}
		if hash.Valid {
			m.Hash = hash.String
		}
		batch = append(batch, m)
		nextID = rowID
	}
	if err = rows.Err(); err != nil {
		return nil, lastID, fmt.Errorf("iterating rows: %w", err)
	}
	return batch, nextID, nil
}
func (h *Handler) Write(ctx context.Context, metric *metrics.Metrics) error {
	err := h.Upsert(ctx, metric)
	if err != nil {
		return fmt.Errorf("upsert metrics: %w", err)
	}
	return nil
}

func (h *Handler) WriteBatch(ctx context.Context, metrics []metrics.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}
	tx, err := h.dbConn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	for i := 0; i < len(metrics); i += batchSize {
		end := i + batchSize
		if end > len(metrics) {
			end = len(metrics)
		}
		batch := metrics[i:end]
		placeholders := make([]string, 0, len(batch))
		args := make([]any, 0, len(batch)*5)
		for j := 0; j < len(batch); j++ {
			base := j*5 + 1
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", base, base+1, base+2, base+3, base+4))
			m := &batch[j]
			args = append(args, m.ID, m.MType, m.Delta, m.Value, m.Hash)
		}
		sqlQuery := "INSERT INTO metrics (name, mtype, delta, value, hash) VALUES " +
			strings.Join(placeholders, ", ") +
			" ON CONFLICT (name) DO UPDATE SET mtype = EXCLUDED.mtype, delta = EXCLUDED.delta, value = EXCLUDED.value, hash = EXCLUDED.hash"
		_, err := tx.ExecContext(ctx, sqlQuery, args...)
		if err != nil {
			errRollback := tx.Rollback()
			if errRollback != nil {
				return fmt.Errorf("failed to rollback: %w", err)
			}
			return fmt.Errorf("failed to exec batch: %w", err)
		}
	}
	return tx.Commit()
}
