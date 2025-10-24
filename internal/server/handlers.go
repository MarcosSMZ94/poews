package server

import (
	"net/http"

	"github.com/coder/websocket"
)

func (app *application) broadcastHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		app.logger.Error("Failed to accept connection", "error", err)
		return
	}

	app.register <- conn

	defer func() {
		app.unregister <- conn
	}()

	ctx := r.Context()
	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 {
				app.logger.Debug("Connection closed normally")
			} else {
				app.logger.Error("Read error", "error", err)
			}
			return
		}
	}
}
