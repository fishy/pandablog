package logging

import (
	"os"

	"go.yhsif.com/ctxslog"
	"golang.org/x/exp/slog"
)

func InitJSON() {
	slog.SetDefault(slog.New(ctxslog.ContextHandler(slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: ctxslog.ChainReplaceAttr(
			ctxslog.GCPKeys,
			ctxslog.StringDuration,
		),
	}.NewJSONHandler(os.Stderr))))
}

func InitText() {
	slog.SetDefault(slog.New(ctxslog.ContextHandler(slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: ctxslog.ChainReplaceAttr(
			ctxslog.StringDuration,
		),
	}.NewTextHandler(os.Stderr))))
}
