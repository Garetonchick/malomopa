package executor

import (
	"malomopa/internal/common"
	"malomopa/internal/config"
	"net/http"
)

type Server struct {
	executorConfig *config.OrderExecutorConfig

	mux *http.ServeMux

	dsProvider common.CacheServiceProvider
	dbProvider common.DBProvider
}

func (s *Server) acquireOrderHandler(w http.ResponseWriter, r *http.Request) {
	// ArtNext
}

func (s *Server) setupRoutes() {
	s.mux = http.NewServeMux()

	s.mux.HandleFunc("POST /v1/acquire_order", s.acquireOrderHandler)
}

func NewServer(cfg *config.OrderExecutorConfig) (*Server, error) {
	server := &Server{
		executorConfig: cfg,
	}

	server.setupRoutes()

	return server, nil
}

func (s *Server) Run() error {
	// ArtNext
	return nil
}
