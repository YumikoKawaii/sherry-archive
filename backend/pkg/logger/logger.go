// Package logger initialises a global zap logger with JSON output.
// Call Init once at startup; use zap.L() / zap.S() everywhere else.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init builds a production JSON logger and installs it as the zap global.
// Returns a flush function that should be deferred in main.
func Init() func() {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.MessageKey = "msg"

	log, err := cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		panic("logger init: " + err.Error())
	}
	zap.ReplaceGlobals(log)
	return func() { _ = log.Sync() }
}
