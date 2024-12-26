package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	log *slog.Logger
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

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	return &Logger{
		log: slog.New(handler),
	}
}

func (l *Logger) Debug(msg string, attrs ...any) {
	l.log.Debug(msg, attrs...)
}

func (l *Logger) Info(msg string, attrs ...any) {
	l.log.Info(msg, attrs...)
}

func (l *Logger) Warn(msg string, attrs ...any) {
	l.log.Warn(msg, attrs...)
}

func (l *Logger) Error(msg string, attrs ...any) {
	l.log.Error(msg, attrs...)
}

func (l *Logger) WithAttrs(attrs ...slog.Attr) *Logger {
	args := make([]any, len(attrs)*2)
	for i, attr := range attrs {
		args[i*2] = attr.Key
		args[i*2+1] = attr.Value
	}
	return &Logger{
		log: l.log.With(args...),
	}
}
