package logger

import (
	"errors"
	"fmt"
	"os"
	"sys-metrics/internal/common"

	"go.uber.org/zap"
)

const logDir = "logs/"

func Initialize(mode, cmdType string) (*zap.Logger, error) {
	var err error
	var log *zap.Logger
	if mode == common.TypeModeProduction {
		log, err = createProductionLogger(cmdType)
	} else if mode == common.TypeModeDevelopment || mode == common.TypeModeTest {
		log, err = createDevelopmentLogger()
	} else {
		err = errors.New("invalid mode")
	}
	if err != nil {
		return nil, err
	}

	return log, nil
}
func createProductionLogger(cmdType string) (*zap.Logger, error) {
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("could not create log directory: %w", err)
	}
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{logDir + cmdType + "_info.log", "stdout"}
	config.ErrorOutputPaths = []string{logDir + cmdType + "_errors.log", "stderr"}
	logger, err := config.Build()
	return logger, err
}
func createDevelopmentLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("could not create development logger: %w", err)
	}
	return logger, nil
}
