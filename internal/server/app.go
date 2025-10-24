package server

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/coder/websocket"
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
			broadcast:  make(chan []byte),
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
		err := conn.Write(context.Background(), websocket.MessageText, []byte(message))
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

func (app *application) ReadInput() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\nType messages to send to clients:")
	for scanner.Scan() {
		message := scanner.Text()
		if message != "" {
			app.Broadcast([]byte(message))
		}
	}
}

func (app *application) Broadcast(message []byte) {
	app.broadcast <- message
}
