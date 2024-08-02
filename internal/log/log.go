package log

import (
	"log/slog"

	"github.com/go-logr/logr"
)

var (
	defaultLevel slog.LevelVar

	Logger = logr.FromSlogHandler(newSLogHandler())
)

func SetLevel(lvl int) {
	defaultLevel.Set(slog.Level(-lvl))
}

func Info(msg string, keysAndValues ...any) {
	Logger.Info(msg, keysAndValues...)
}

func Error(err error, msg string, keysAndValues ...any) {
	Logger.Error(err, msg, keysAndValues...)
}

func V(level int) logr.Logger {
	return Logger.V(level)
}
