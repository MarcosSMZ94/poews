package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 60 * time.Second
)

type SocketClient struct {
	clients   map[*websocket.Conn]bool
	clientMu  sync.Mutex
	broadcast chan string
}

type application struct {
	*SocketClient
	logger *slog.Logger
}

func (app *application) broadcastHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		app.logger.Error("Failed to accept connection", "error", err)
		return
	}

	defer func() {
		if err := conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			app.logger.Error("Error closing server", "error", err)
		}
	}()

	app.clientMu.Lock()
	app.clients[conn] = true
	totalClients := len(app.clients)
	app.clientMu.Unlock()
	app.logger.Info("New client connected", "total_clients", totalClients)

	defer func() {
		app.clientMu.Lock()
		delete(app.clients, conn)
		totalClients := len(app.clients)
		app.clientMu.Unlock()
		app.logger.Info("Client disconnected", "total_clients", totalClients)
	}()

	ctx := r.Context()
	for {
		_, _, err := conn.Read(ctx)
		if websocket.CloseStatus(err) != -1 {
			app.logger.Warn("Connection closed", "error", err)
			return
		}
		if err != nil {
			app.logger.Error("Read error", "error", err)
			return
		}
	}
}

func (app *application) broadcastMessage() {
	ctx := context.Background()

	for message := range app.broadcast {
		app.clientMu.Lock()
		for conn := range app.clients {
			err := conn.Write(ctx, websocket.MessageText, []byte(message))
			if err != nil {
				app.logger.Error("Error sending to client", "error", err)
				if err := conn.Close(websocket.StatusInternalError, ""); err != nil {
					app.logger.Error("Error closing client connection", "error", err)
				}
				delete(app.clients, conn)
			}
		}
		totalClients := len(app.clients)
		app.clientMu.Unlock()
		app.logger.Info("Broadcasted message", "total_clients", totalClients, "message", message)
	}
}

func (app *application) readInput() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\nType messages to send to clients:")
	for scanner.Scan() {
		message := scanner.Text()
		if message != "" {
			app.broadcast <- message
		}
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := &application{
		SocketClient: &SocketClient{
			clients:   make(map[*websocket.Conn]bool),
			clientMu:  sync.Mutex{},
			broadcast: make(chan string),
		},
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", app.broadcastHandler)

	go app.broadcastMessage()
	go app.readInput()

	app.logger.Info("Server started at ws://localhost:8080/ws")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	err := server.ListenAndServe()
	log.Fatal(err)
}
