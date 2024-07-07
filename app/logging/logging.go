package logging

import (
	"log/slog"
	"os"

	"go.yhsif.com/ctxslog"
)

func InitJSON(level slog.Level) {
	logger := ctxslog.New(
		ctxslog.WithAddSource(true),
		ctxslog.WithLevel(level),
		ctxslog.WithCallstack(slog.LevelError),
		ctxslog.WithReplaceAttr(ctxslog.ChainReplaceAttr(
			ctxslog.GCPKeys,
			ctxslog.StringDuration,
		)),
	)
	if v, ok := os.LookupEnv("VERSION_TAG"); ok {
		logger = logger.With(slog.String("v", v))
	}
	slog.SetDefault(logger)
}

func InitText(level slog.Level) {
	slog.SetDefault(ctxslog.New(
		ctxslog.WithText,
		ctxslog.WithAddSource(true),
		ctxslog.WithLevel(level),
		ctxslog.WithCallstack(slog.LevelError),
		ctxslog.WithReplaceAttr(ctxslog.ChainReplaceAttr(
			ctxslog.StringDuration,
		)),
	))
}
