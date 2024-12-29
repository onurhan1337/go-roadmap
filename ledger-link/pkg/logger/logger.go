package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func New(level string) *Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: logLevel}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{logger}
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Logger.Error(msg, args...)
	os.Exit(1)
}

func (l *Logger) Debug(msg string, attrs ...any) {
	l.Logger.Debug(msg, attrs...)
}

func (l *Logger) Info(msg string, attrs ...any) {
	l.Logger.Info(msg, attrs...)
}

func (l *Logger) Warn(msg string, attrs ...any) {
	l.Logger.Warn(msg, attrs...)
}

func (l *Logger) Error(msg string, attrs ...any) {
	l.Logger.Error(msg, attrs...)
}

func (l *Logger) WithAttrs(attrs ...slog.Attr) *Logger {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return &Logger{l.With(args...)}
}
