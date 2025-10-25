package main

import (
	"log/slog"
	"os"
	"path/filepath"
)

func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String("time", a.Value.Time().Format(dateTimeFormat))
			}

			if a.Key == "path" || a.Key == "watching" {
				if str, ok := a.Value.Any().(string); ok {
					return slog.String(a.Key, filepath.ToSlash(str))
				}
			}

			return a
		},
	}))
}
