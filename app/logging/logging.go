package logging

import (
	"os"

	"golang.org/x/exp/slog"
)

func renderValues(v slog.Value) slog.Value {
	switch v.Kind() {
	default:
		return v
	case slog.KindDuration:
		return slog.StringValue(v.Duration().String())
	}
}

func InitJSON() {
	slog.SetDefault(slog.New(slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 {
				switch a.Key {
				case slog.MessageKey:
					a.Key = "message"
				case slog.LevelKey:
					a.Key = "severity"
				}
			}
			a.Value = renderValues(a.Value)
			return a
		},
	}.NewJSONHandler(os.Stderr)))
}

func InitText() {
	slog.SetDefault(slog.New(slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			a.Value = renderValues(a.Value)
			return a
		},
	}.NewTextHandler(os.Stderr)))
}
