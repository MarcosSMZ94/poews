package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func setupSignalHandler(shutdown chan struct{}, logger *slog.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		close(shutdown)
	}()
}
