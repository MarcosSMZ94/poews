package main

import (
	"context"
	"log/slog"
	"os/exec"
	"runtime"

	"github.com/MarcosSMZ94/poews/internal/server"
	"github.com/getlantern/systray"
)

func runSystemTray(srv *server.Server, logger *slog.Logger, addr string, shutdown chan struct{}) {
	systray.Run(
		func() {
			onSystrayReady(logger, addr, shutdown)
		},
		func() {
			onSystrayExit(srv, logger)
		},
	)
}

func onSystrayReady(logger *slog.Logger, addr string, shutdown chan struct{}) {
	systray.SetTitle("PoE Trade Helper")
	systray.SetTooltip("PoE Trade Helper - Monitoring trades")

	mOpen := systray.AddMenuItem("Open Config", "Open configuration page")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the app")

	go handleSystrayEvents(mOpen, mQuit, logger, addr, shutdown)
}

func handleSystrayEvents(mOpen, mQuit *systray.MenuItem, logger *slog.Logger, addr string, shutdown chan struct{}) {
	for {
		select {
		case <-mOpen.ClickedCh:
			openConfig(logger, addr)
		case <-mQuit.ClickedCh:
			logger.Info("Shutting down gracefully...")
			close(shutdown)
			systray.Quit()
			return
		case <-shutdown:
			systray.Quit()
			return
		}
	}
}

func openURL(url string) error {
	ctx := context.Background()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", url)
	case "linux":
		cmd = exec.CommandContext(ctx, "xdg-open", url)
	default:
		return nil
	}
	return cmd.Start()
}

func openConfig(logger *slog.Logger, addr string) {
	url := "http://localhost" + addr + "/config"
	if err := openURL(url); err != nil {
		logger.Error("Failed to open config", "error", err)
	}
}

func onSystrayExit(srv *server.Server, logger *slog.Logger) {
	logger.Info("Cleaning up...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Shutdown error", "error", err)
	}

	logger.Info("Shutdown complete")
}
