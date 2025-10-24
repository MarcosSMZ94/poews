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
	pollInterval        = 1000 * time.Millisecond
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

	logger        *slog.Logger
	watchedFile   string
	watchedFileMu sync.RWMutex
	fileOffset    int64
	fileOffsetMu  sync.Mutex
	stopWatcher   chan struct{}
	stopWatcherMu sync.Mutex
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
	defer app.mu.Unlock()

	if _, ok := app.clients[conn]; ok {
		delete(app.clients, conn)
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}

	app.logger.Info("Client disconnected", "total", len(app.clients))
}

func (app *application) handleBroadcast(message []byte) {
	app.mu.Lock()
	defer app.mu.Unlock()

	for conn := range app.clients {
		if err := conn.Write(context.Background(), websocket.MessageText, message); err != nil {
			app.logger.Error("Error sending to client", "error", err)
			_ = conn.Close(websocket.StatusInternalError, "")
			delete(app.clients, conn)
		}
	}

	app.logger.Info("Broadcasted message", "total_clients", len(app.clients), "message", message)
}

func (app *application) sendMessage(message []byte) {
	select {
	case app.broadcast <- message:
		//  Success
	default:
		app.logger.Warn("Broadcast channel full, message dropped")
	}
}

func (app *application) setWatchedFile(filepath string) {
	app.watchedFileMu.Lock()
	defer app.watchedFileMu.Unlock()
	app.watchedFile = filepath
}

func (app *application) getWatchedFile() string {
	app.watchedFileMu.RLock()
	defer app.watchedFileMu.RUnlock()
	return app.watchedFile
}

func (app *application) setFileOffset(offset int64) {
	app.fileOffsetMu.Lock()
	defer app.fileOffsetMu.Unlock()
	app.fileOffset = offset
}

func (app *application) getFileOffset() int64 {
	app.fileOffsetMu.Lock()
	defer app.fileOffsetMu.Unlock()
	return app.fileOffset
}

func (app *application) WatchFile(filepath string) {
	app.setWatchedFile(filepath)

	if offset, err := utils.GetCurrentFileSize(filepath); err == nil {
		app.setFileOffset(offset)
		app.logger.Info("File watcher initialized", "path", filepath)
	} else {
		app.logger.Error("Failed to get initial file size", "error", err)
		return
	}

	// Prevent multiples goroutines for watchfile
	app.stopWatcherMu.Lock()
	if app.stopWatcher != nil {
		close(app.stopWatcher)
	}
	app.stopWatcher = make(chan struct{})
	stopChan := app.stopWatcher
	app.stopWatcherMu.Unlock()

	ticker := time.NewTicker(pollInterval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				app.processNewMessages()
			case <-stopChan:
				app.logger.Info("File watcher stopped")
				return
			}
		}
	}()
}

func (app *application) processNewMessages() {
	filepath := app.getWatchedFile()
	currentOffset := app.getFileOffset()

	messages, newOffset, err := utils.GetNewTradeMessages(filepath, currentOffset)
	if err != nil {
		app.logger.Error("Failed to read messages", "error", err, "offset", currentOffset)
		return
	}

	if len(messages) == 0 {
		return
	}

	for _, message := range messages {
		app.sendMessage([]byte(message))
	}

	app.setFileOffset(newOffset)
}

func (app *application) stopFileWatcher() {
	app.stopWatcherMu.Lock()
	defer app.stopWatcherMu.Unlock()

	if app.stopWatcher != nil {
		close(app.stopWatcher)
		app.stopWatcher = nil
	}
}

func (app *application) closeAllClients() {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.logger.Info("Closing all WebSocket connections", "total", len(app.clients))

	for conn := range app.clients {
		_ = conn.Close(websocket.StatusGoingAway, "Server shutting down")
		delete(app.clients, conn)
	}
}
