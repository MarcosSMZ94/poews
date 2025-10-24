package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/MarcosSMZ94/poews/internal/server"
)


func main() {
	addr := flag.String("addr", ":8080", "HTTP network address")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := server.NewServer(*addr, logger)

	go server.ReadInput()

	err := server.Start()
	log.Fatal(err)
}
