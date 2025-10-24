package server

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/MarcosSMZ94/poews/internal/utils"
	"github.com/coder/websocket"
)

const (
	pollInterval        = 500 * time.Millisecond
	broadcastBufferSize = 256
)

type socketClient struct {
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan []byte
	mu         sync.Mutex
}

type application struct {
	*socketClient

	logger *slog.Logger
}

func newApp(logger *slog.Logger) *application {
	return &application{
		socketClient: &socketClient{
			clients:    make(map[*websocket.Conn]bool),
			register:   make(chan *websocket.Conn),
			unregister: make(chan *websocket.Conn),
			broadcast:  make(chan []byte, broadcastBufferSize),
			mu:         sync.Mutex{},
		},
		logger: logger,
	}
}

func (app *application) run() {
	for {
		select {
		case conn := <-app.register:
			app.handleRegister(conn)

		case conn := <-app.unregister:
			app.handleUnregister(conn)

		case message := <-app.broadcast:
			app.handleBroadcast(message)
		}
	}
}

func (app *application) handleRegister(conn *websocket.Conn) {
	app.mu.Lock()
	app.clients[conn] = true
	total := len(app.clients)
	app.mu.Unlock()

	app.logger.Info("Client connected", "total", total)
}

func (app *application) handleUnregister(conn *websocket.Conn) {
	app.mu.Lock()
	if _, ok := app.clients[conn]; ok {
		delete(app.clients, conn)
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}
	total := len(app.clients)
	app.mu.Unlock()

	app.logger.Info("Client disconnected", "total", total)
}

func (app *application) handleBroadcast(message []byte) {
	app.mu.Lock()
	for conn := range app.clients {
		err := conn.Write(context.Background(), websocket.MessageText, message)
		if err != nil {
			app.logger.Error("Error sending to client", "error", err)
			if err := conn.Close(websocket.StatusInternalError, ""); err != nil {
				app.logger.Error("Error closing client connection", "error", err)
			}
			delete(app.clients, conn)
		}
	}
	app.mu.Unlock()
	app.logger.Info("Broadcasted message", "total_clients", len(app.clients), "message", message)
}

func (app *application) WatchFile(filepath string) {
	lastOffset, err := utils.GetCurrentFileSize(filepath)
	if err != nil {
		app.logger.Error("Failed to get initial file size", "error", err)
		lastOffset = 0
	}

	ticker := time.NewTicker(pollInterval)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			messages, newOffset, err := utils.GetNewTradeMessages(filepath, lastOffset)
			if err != nil {
				app.logger.Error("Failed to read messages", "error", err)
				continue
			}

			if len(messages) > 0 {
				for _, message := range messages {
					app.broadcast <- []byte(message)
				}
				lastOffset = newOffset
			}
		}
	}()
}
