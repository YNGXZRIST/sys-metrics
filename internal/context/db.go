package context

import (
	"context"
	"errors"
	"sys-metrics/internal/common"
	"sys-metrics/internal/config/db"
)

func DBFromContext(ctx context.Context) (*db.DB, error) {
	conn, ok := ctx.Value(common.ContextDBKey).(*db.DB)
	if !ok {
		return nil, errors.New("db context not found in context")
	}
	return conn, nil
}
