package logger

import (
	"errors"
	"sys-metrics/internal/common"
	"sys-metrics/pkg/filesystem"

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
	err := filesystem.CreateDirIfNotExists(logDir)
	if err != nil {
		return nil, err
	}
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{logDir + cmdType + "_info.log", "stdout"}
	config.ErrorOutputPaths = []string{logDir + cmdType + "_errors.log", "stderr"}
	logger, err := config.Build()
	return logger, err
}
func createDevelopmentLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	return logger, err
}
