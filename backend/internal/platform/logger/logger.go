package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

func New(cfg config.LoggingConfig) *slog.Logger {
	return NewWithWriter(cfg, os.Stdout)
}

func NewWithWriter(cfg config.LoggingConfig, writer io.Writer) *slog.Logger {
	level := slog.LevelInfo
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(writer, opts)
	} else {
		handler = slog.NewJSONHandler(writer, opts)
	}

	return slog.New(handler)
}
