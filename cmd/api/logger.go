package main

import (
	"log/slog"
	"os"
)

func setupLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format(dateTimeFormat))
			}
			return a
		},
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}
