package logger

import (
	"io"
	"log/slog"
	"os"

	"zsxyww.com/wts/config"
)

func Setup(cfg *config.Config) {

	var HumanHandler slog.Handler
	var JSONHandler slog.Handler
	var level slog.Level
	var Logger *slog.Logger

	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	slog.SetLogLoggerLevel(level)

	// TODO: 外部收集JSON输出
	HumanHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	JSONHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})

	if cfg.JSONLogOutput {
		Logger = slog.New(JSONHandler)
	} else {
		Logger = slog.New(HumanHandler)
	}

	slog.SetDefault(Logger)
}

type Human struct {
	slog.Handler
	writer io.Writer
	level  slog.Leveler
}
