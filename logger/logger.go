package logger

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func InitSlogLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}
