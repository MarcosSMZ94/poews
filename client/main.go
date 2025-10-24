package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/coder/websocket"
)

func main() {
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, "ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "Client closing connection")

	fmt.Println("Listening for messages")

	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		for scanner.Scan() {
			if scanner.Text() == "quit" {
				conn.Close(websocket.StatusNormalClosure, "Client requested close")
				os.Exit(0)
			}
		}
	}()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			log.Printf("Read error: %v", err)
			return
		}
		fmt.Printf("Received: %s\n", data)
	}
}
