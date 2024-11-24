package executor

import (
	"errors"
	"fmt"
	"malomopa/internal/common"
	"malomopa/internal/config"
	"malomopa/internal/db"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	*common.MonitoringServer

	executorConfig *config.OrderExecutorConfig

	mux *chi.Mux

	dbProvider common.DBProvider
}

const (
	executorServiceName = "executor"
)

func (s *Server) acquireOrderHandler(w http.ResponseWriter, r *http.Request) {
	common.AddRequest()

	executorID := common.FetchQueryParam(r, common.ExecutorIDQueryParam)
	handlerCtx := r.Context()
	logger := common.GetRequestLogger(handlerCtx, executorServiceName, "acquire_order")

	if executorID == nil {
		logger.Error("not all query params supplied",
			zap.Bool("executor_id_is_nil", executorID == nil),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := s.dbProvider.AcquireOrder(handlerCtx, *executorID)
	if err != nil {
		logger.Error("failed to acquire order",
			zap.String("executor_id", *executorID),
		)
		if errors.Is(err, db.ErrOrderNotFound) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Write(payload)

	logger.Info("request to acquire order is processed",
		zap.String("executor_id", *executorID),
	)
}

func (s *Server) setupRoutes(logger *zap.Logger) {
	s.mux = chi.NewRouter()

	common.SetupMiddlewares(s.mux, logger)
	s.mux.Post("/v1/acquire_order", s.acquireOrderHandler)
	s.MonitoringServer.SetupRoutes(s.mux)
}

func NewServer(cfg *config.OrderExecutorConfig, dbProvider common.DBProvider, logger *zap.Logger) (*Server, error) {
	server := &Server{
		MonitoringServer: &common.MonitoringServer{},
		executorConfig:   cfg,
		dbProvider:       dbProvider,
	}

	server.setupRoutes(logger)

	return server, nil
}

func (s *Server) Run(logger *zap.Logger) error {
	logger.Info("Starting HTTP Server...")
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.executorConfig.HTTPServer.Host, s.executorConfig.HTTPServer.Port), s.mux)
}
