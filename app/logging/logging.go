package logging

import (
	"os"

	"go.yhsif.com/ctxslog"
	"golang.org/x/exp/slog"
)

func InitJSON() {
	slog.SetDefault(slog.New(ctxslog.ContextHandler(ctxslog.JSONCallstackHandler(
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
	))))
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
