package logger

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger{
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // minimum log level
	})

	logger := slog.New(handler)

	return logger
}