package logging

import (
	"os"

	"go.yhsif.com/ctxslog"
	"golang.org/x/exp/slog"
)

func InitJSON() {
	logger := slog.New(ctxslog.ContextHandler(ctxslog.JSONCallstackHandler(
		slog.NewJSONHandler(
			os.Stderr,
			&slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
				ReplaceAttr: ctxslog.ChainReplaceAttr(
					ctxslog.GCPKeys,
					ctxslog.StringDuration,
				),
			},
		),
		slog.LevelError,
	)))
	if v, ok := os.LookupEnv("VERSION_TAG"); ok {
		logger = logger.With(slog.String("v", v))
	}
	slog.SetDefault(logger)
}

func InitText() {
	slog.SetDefault(slog.New(ctxslog.ContextHandler(ctxslog.TextCallstackHandler(
		slog.NewTextHandler(
			os.Stderr,
			&slog.HandlerOptions{
				AddSource: true,
				Level:     slog.LevelDebug,
				ReplaceAttr: ctxslog.ChainReplaceAttr(
					ctxslog.StringDuration,
				),
			},
		),
		slog.LevelError,
	))))
}
