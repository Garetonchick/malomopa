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
	executorConfig *config.OrderExecutorConfig

	mux *chi.Mux

	dbProvider common.DBProvider
}

const (
	system = "acquirer"
)

func (s *Server) acquireOrderHandler(w http.ResponseWriter, r *http.Request) {
	executorID := common.FetchQueryParam(r, common.ExecutorIDQueryParam)
	handlerCtx := r.Context()
	logger := common.GetRequestLogger(handlerCtx, system, "acquire_order")

	if executorID == nil {
		logger.Error("not all query params supplied",
			zap.Bool("order_id_is_nil", executorID == nil),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := s.dbProvider.AcquireOrder(handlerCtx, *executorID)
	if err != nil {
		logger.Error("failed to cancel order",
			zap.String("order_id", *executorID),
		)
		if errors.Is(err, db.ErrNoSuchRowToUpdate) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Write(payload)

	logger.Info("request to cancel order is processed",
		zap.String("executor_id", *executorID),
	)
}

func (s *Server) setupRoutes(logger *zap.Logger) {
	s.mux = chi.NewRouter()

	common.SetupMiddlewares(s.mux, logger)
	s.mux.Post("/v1/acquire_order", s.acquireOrderHandler)
}

func NewServer(cfg *config.OrderExecutorConfig, dbProvider common.DBProvider, logger *zap.Logger) (*Server, error) {
	server := &Server{
		executorConfig: cfg,
		dbProvider:     dbProvider,
	}

	server.setupRoutes(logger)

	return server, nil
}

func (s *Server) Run() error {
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.executorConfig.HTTPServer.Host, s.executorConfig.HTTPServer.Port), s.mux)
}
