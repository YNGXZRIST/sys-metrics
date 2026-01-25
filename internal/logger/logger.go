package logger

import (
	"errors"
	"log"
	"sys-metrics/internal/common"

	"go.uber.org/zap"
)

func Initialize(mode string) (*log.Logger, error) {
	var logger *zap.Logger
	var err error
	if mode == common.TypeModeProduction {
		logger, err = zap.NewProduction()
	} else if mode == common.TypeModeDevelopment {
		logger, err = zap.NewDevelopment()
	} else {
		err = errors.New("invalid mode")
	}
	if err != nil {
		return nil, err
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Printf("failed to sync logger: %v", err)
		}
	}(logger)
	return zap.NewStdLog(logger), nil
}
