package postgress

import (
	"context"
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
	var excludedField string
	switch metric.MType {
	case common.Gauge:
		excludedField = "value"
	case common.Counter:
		excludedField = "delta"
	default:
		return fmt.Errorf("unsupported metric type: %s", metric.MType)

	}
	sqlStatement := `
    INSERT INTO metrics (id, mtype, delta,value,hash)
    VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT (id)
    DO UPDATE SET
        $6 = EXCLUDED.$6,`
	_, err := h.dbConn.ExecContext(ctx, sqlStatement, metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash, excludedField)
	if err != nil {
		return fmt.Errorf("upsert failed: %w", err)
	}
	panic("implement me")
}

func (h Handler) Read(ctx context.Context) ([]metrics.Metrics, error) {
	//TODO implement me
	panic("implement me")
}

func (h Handler) Write(ctx context.Context, metric *metrics.Metrics) error {
	//TODO implement me
	panic("implement me")
}

func (h Handler) WriteBatch(ctx context.Context, metrics []metrics.Metrics) error {
	//TODO implement me
	panic("implement me")
}
