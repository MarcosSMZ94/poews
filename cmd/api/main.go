package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/MarcosSMZ94/poews/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP network address")
	filepath := flag.String(
		"path",
		"C:\\Program Files (x86)\\Steam\\steamapps\\common\\Path of Exile\\logs\\LatestClient.txt",
		"PoE default path",
	)
	flag.Parse()

	opts := &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format("02/01/2006 15:04:05"))
			}
			return a
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	server := server.NewServer(*addr, logger)
	server.WatchFile(*filepath)

	err := server.Start()
	log.Fatal(err)
}
