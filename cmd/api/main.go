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
	defaultFilePath = `C:\Program Files (x86)\Steam\steamapps\common\Path of Exile\logs\LatestClient.txt`
	dateTimeFormat  = "02/01/2006 15:04:05"
	shutdownTimeout = 5 * time.Second
)

func main() {
	LoadEnv()

	addr := flag.String("addr", getEnvOrDefault("SERVER_ADDR", defaultAddr),
		"HTTP server address (env: SERVER_ADDR")

	filepath := flag.String("path", getEnvOrDefault("POE_LOG_PATH", defaultFilePath),
		"Path to PoE Client.txt log file (env: POE_LOG_PATH)")

	resetConfig := flag.Bool("reset", false,
		"Remove .env file to reset configuration to defaults")

	flag.Parse()

	if *resetConfig {
		if err := resetToDefaults(); err != nil {
			log.Fatalf("Failed to reset config: %v", err)
		}
		log.Println("Configuration reset: .env file cleared")
		return
	}

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
