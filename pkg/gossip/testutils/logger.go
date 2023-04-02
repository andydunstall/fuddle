package testutils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Logger() *zap.Logger {
	logLevel := os.Getenv("FUDDLE_LOG_LEVEL")

	loggerConf := zap.NewProductionConfig()
	switch logLevel {
	case "debug":
		loggerConf.Level.SetLevel(zapcore.DebugLevel)
	case "info":
		loggerConf.Level.SetLevel(zapcore.InfoLevel)
	case "warn":
		loggerConf.Level.SetLevel(zapcore.WarnLevel)
	case "error":
		loggerConf.Level.SetLevel(zapcore.ErrorLevel)
	default:
		// If the level is invalid or not specified don't use a logger.
		return zap.NewNop()
	}
	return zap.Must(loggerConf.Build())
}
