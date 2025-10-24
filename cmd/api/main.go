package main

import (
	"flag"
	"log"
	"log/slog"
	"time"

	"github.com/MarcosSMZ94/poews/internal/server"
)

const (
	defaultAddr     = ":8080"
	defaultFilePath = "C:\\Program Files (x86)\\Steam\\steamapps\\common\\Path of Exile\\logs\\LatestClient.txt"
	dateTimeFormat  = "02/01/2006 15:04:05"
	shutdownTimeout = 5 * time.Second
)

func main() {
	addr := flag.String("addr", defaultAddr, "HTTP network address")
	filepath := flag.String("path", defaultFilePath, "PoE client log path")
	flag.Parse()

	if err := validateFilePath(*filepath); err != nil {
		log.Fatalf("Invalid file path: %v", err)
	}

	logger := setupLogger()
	logger.Info("Starting PoE Trade Helper", "addr", *addr, "watching", *filepath)

	srv := server.NewServer(*addr, logger)
	srv.WatchFile(*filepath)

	shutdown := make(chan struct{})
	setupSignalHandler(shutdown, logger)

	startServer(srv, logger)

	runSystemTray(srv, logger, *addr, shutdown)
}

func startServer(srv *server.Server, logger *slog.Logger) {
	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("Server error", "error", err)
		}
	}()
}
