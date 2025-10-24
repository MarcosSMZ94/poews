package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *server) registerRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.HandleFunc("/ws", s.broadcastHandler)

	return r
}
