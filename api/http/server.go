package http

import (
	"context"
	"net"
	"net/http"
	"time"

	config "github.com/elkoshar/reconciliation-app/configs"
)

// Server struct
type Server struct {
	server *http.Server
	Cfg    *config.Config
}

var ()

// Serve will run an HTTP server
func (s *Server) Serve(port string) error {

	s.server = &http.Server{
		ReadTimeout:  s.Cfg.HttpReadTimeout * time.Second,
		WriteTimeout: s.Cfg.HttpWriteTimeout * time.Second,
		Handler:      handler(s.Cfg),
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	return s.server.Serve(lis)
}

// Shutdown will tear down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
