package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Initialize sets up the logger
func Initialize(isDevelopment bool) error {
	var cfg zap.Config
	if isDevelopment {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}

	var err error
	log, err = cfg.Build()
	if err != nil {
		return err
	}

	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if log == nil {
		log, _ = zap.NewProduction()
	}
	return log
}

// Sync flushes any buffered log entries
func Sync() error {
	return log.Sync()
} 