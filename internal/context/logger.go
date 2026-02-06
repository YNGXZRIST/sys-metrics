package context

import (
	"context"
	"sys-metrics/internal/common"

	"go.uber.org/zap"
)

func LoggerFromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(common.ContextLoggerKey).(*zap.Logger)
	if !ok {
		return zap.NewNop()
	}
	return logger
}
