package config

import (
	"log/slog"

	"github.com/m-mizutani/clog"
)

var logLevel slog.Leveler = slog.LevelInfo

func logHandler(level slog.Leveler) *clog.Handler {
	return clog.New(
		clog.WithColor(true),
		clog.WithSource(true),
		clog.WithLevel(level),
	)
}

func logConfig() {
	logger := slog.New(logHandler(logLevel))
	slog.SetDefault(logger)
}
