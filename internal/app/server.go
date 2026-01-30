package app

import (
	"context"
	"net/http"
	"wb-project/internal/handler"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(orderHandler *handler.OrderHandler) *Server {
	router := handler.NewRouter(orderHandler)

	return &Server{
		httpServer: &http.Server{
			Handler: router,
		},
	}
}

func (s *Server) Run(addr string) error {
	s.httpServer.Addr = addr
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
