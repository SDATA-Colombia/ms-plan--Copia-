package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

func Init(environment string) {
	level := slog.LevelInfo
	isDev := environment == "development" || environment == "dev"

	if isDev {
		level = slog.LevelDebug
	}

	var handler slog.Handler
	if isDev {
		// Formato legible para humanos en tu Debian local
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		// Formato JSON estructurado para que Victoria Metrics indexe tus facturas en producción
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	log = slog.New(handler)
	slog.SetDefault(log)
}

func Default() *slog.Logger {
	if log != nil {
		return log
	}

	Init("production")
	return log
}

func Info(message string, args ...any) {
	if log != nil {
		log.Info(message, args...)
	}
}

func Error(message string, args ...any) {
	if log != nil {
		log.Error(message, args...)
	}
}

func Warn(message string, args ...any) {
	if log != nil {
		log.Warn(message, args...)
	}
}

func Debug(message string, args ...any) {
	if log != nil {
		log.Debug(message, args...)
	}
}

func Fatal(message string) {
	if log != nil {
		log.Error(message)
	}
	os.Exit(1)
}
