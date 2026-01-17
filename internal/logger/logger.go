package logger

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	local = "local"
	prod  = "prod"
	dev   = "dev"
)

func New(env string) *slog.Logger {
	logLvl := logLvl(env)

	var logger *slog.Logger
	switch env {
	case local:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     logLvl,
			AddSource: true,
		}))
	case prod:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     logLvl,
			AddSource: true,
		}))
	case dev:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     logLvl,
			AddSource: false,
		}))
	default:
		panic(fmt.Sprintf("unexpected env: %s", env))
	}

	return logger
}

func logLvl(env string) slog.Level {
	switch env {
	case local:
		return slog.LevelDebug
	case prod:
		return slog.LevelInfo
	case dev:
		return slog.LevelDebug
	}

	return slog.LevelDebug
}
