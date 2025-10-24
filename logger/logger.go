package logger

import (
	"log/slog"
	"os"
)

func InitSlog(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}),
	)

	slog.SetDefault(logger)
}
