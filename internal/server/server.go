package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 60 * time.Second
)

type Server struct {
	*application

	logger     *slog.Logger
	port       string
	httpServer *http.Server
}

func NewServer(port string, logger *slog.Logger) *Server {
	app := newApp(logger)

	return &Server{
		application: app,
		logger:      logger,
		port:        port,
	}
}

func (s *Server) Start() error {
	go s.application.run()

	s.httpServer = &http.Server{
		Addr:         s.port,
		Handler:      s.registerRoutes(),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	s.logger.Info("Server starting", "addr", s.port)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.InfoContext(ctx, "Shutting down server...")

	s.application.stopFileWatcher()

	s.application.closeAllClients()

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}
