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

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := server.NewServer(*addr, logger)
	server.WatchFile(*filepath)

	err := server.Start()
	log.Fatal(err)
}
