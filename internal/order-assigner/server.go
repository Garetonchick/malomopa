package assigner

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
	cfg *config.HTTPServerConfig

	mux *chi.Mux

	csProvider     common.CacheServiceProvider
	costCalculator common.CostCalculator
	dbProvider     common.DBProvider
}

func (s *Server) assignOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := common.FetchQueryParam(r, common.OrderIDQueryParam)
	executorID := common.FetchQueryParam(r, common.ExecutorIDQueryParam)
	handlerCtx := r.Context()

	logger := common.GetRequestLogger(handlerCtx)
	if !logger.IsValid() {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if orderID == nil || executorID == nil {
		logger.Error("not all query params supplied",
			zap.Bool("order_id_is_nil", orderID == nil),
			zap.Bool("executor_id_is_nil", executorID == nil),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// orderInfo, err := s.csProvider.GetOrderInfo(handlerCtx, *orderID, *executorID)
	// if err != nil {
	// 	logger.Error("failed to get order info",
	// 		zap.Error(err),
	// 		zap.String("order_id", *orderID),
	// 		zap.String("executor_id", *executorID),
	// 	)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// cost, err := s.costCalculator.CalculateCost(orderInfo)
	// if err != nil {
	// 	logger.Error("failed to calculate cost",
	// 		zap.Error(err),
	// 		zap.Any("order_info", orderInfo),
	// 	)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// payload, err := json.Marshal(orderInfo)
	// if err != nil {
	// 	logger.Error("failed to marshal order info",
	// 		zap.Error(err),
	// 		zap.Any("order_info", orderInfo),
	// 	)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	var cost float32 = 3.22
	payload := []byte("infra")

	order := common.Order{
		OrderID:    *orderID,
		ExecutorID: *executorID,
		Cost:       cost,
		Payload:    payload,
	}

	err := s.dbProvider.CreateOrder(handlerCtx, &order)
	if err != nil {
		logger.Error("failed to create order in DB",
			zap.Error(err),
			zap.String("order_id", *orderID),
			zap.String("executor_id", *executorID),
			zap.Float32("cost", cost),
			zap.String("payload", string(payload)),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Info("request to create order is processed",
		zap.String("order_id", *orderID),
		zap.String("executor_id", *executorID),
	)
}

func (s *Server) cancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := common.FetchQueryParam(r, common.OrderIDQueryParam)
	handlerCtx := r.Context()

	logger := common.GetRequestLogger(handlerCtx)
	if !logger.IsValid() {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if orderID == nil {
		logger.Error("not all query params supplied",
			zap.Bool("order_id_is_nil", orderID == nil),
		)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := s.dbProvider.CancelOrder(handlerCtx, *orderID)
	if err != nil {
		logger.Error("failed to cancel order",
			zap.Error(err),
			zap.String("order_id", *orderID),
		)
		if errors.Is(err, db.ErrNoSuchRowToUpdate) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// RomGol TODO: add careful error handling here
	w.Write(payload)

	logger.Info("request to cancel order is processed",
		zap.String("order_id", *orderID),
	)
}

func (s *Server) setupRoutes(logger *zap.Logger) {
	s.mux = chi.NewRouter()

	common.SetupMiddlewares(s.mux, logger)
	s.mux.Post("/v1/assign_order", s.assignOrderHandler)
	s.mux.Post("/v1/cancel_order", s.cancelOrderHandler)
}

func NewServer(
	cfg *config.HTTPServerConfig,
	csProvider common.CacheServiceProvider,
	dbProvider common.DBProvider,
	costCalculator common.CostCalculator,
	logger *zap.Logger,
) (*Server, error) {
	server := &Server{
		cfg:            cfg,
		csProvider:     csProvider,
		costCalculator: costCalculator,
		dbProvider:     dbProvider,
	}

	server.setupRoutes(logger)

	return server, nil
}

func (s *Server) Run() error {
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port), s.mux)
}
