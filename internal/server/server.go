package server

import (
	"log/slog"
	"net/http"
	"time"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 60 * time.Second
)

type server struct {
	*application

	logger *slog.Logger
	port   string
}

func NewServer(port string, logger *slog.Logger) *server {
	app := newApp(logger)

	return &server{
		application: app,
		logger:      logger,
		port:        port,
	}
}

func (s *server) Start() error {
	go s.application.run()

	httpServer := &http.Server{
		Addr:         s.port,
		Handler:      s.registerRoutes(),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	s.logger.Info("Server starting", "addr", s.port)
	return httpServer.ListenAndServe()
}
