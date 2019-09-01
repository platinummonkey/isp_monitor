package log

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

// Initialize will initialize the logger.
func Initialize(debug bool) {
	if logger == nil {
		if debug {
			logger, _ = zap.NewDevelopment()
		} else {
			logger, _ = zap.NewProduction()
		}
	}
}

// Get returns the current logger
func Get() *zap.Logger {
	if logger == nil {
		// test interface
		Initialize(true)
	}
	return logger
}
