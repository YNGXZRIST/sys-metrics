package logger

import (
	"errors"
	"sys-metrics/internal/common"
	"sys-metrics/pkg/filesystem"

	"go.uber.org/zap"
)

const logDir = "logs/"

var Log *zap.Logger = zap.NewNop()

func Initialize(mode, cmdType string) (*zap.Logger, error) {
	var err error
	if mode == common.TypeModeProduction {
		Log, err = createProductionLogger(cmdType)
	} else if mode == common.TypeModeDevelopment {
		Log, err = createDevelopmentLogger()
	} else {
		err = errors.New("invalid mode")
	}
	if err != nil {
		return nil, err
	}

	return Log, nil
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
