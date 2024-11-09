package assigner

import (
	"encoding/json"
	"errors"
	"fmt"
	"malomopa/internal/common"
	"malomopa/internal/config"
	"malomopa/internal/db"
	"net/http"

	"github.com/go-chi/chi/v5"
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

	if orderID == nil || executorID == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderInfo, err := s.csProvider.GetOrderInfo(handlerCtx, *orderID, *executorID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cost, err := s.costCalculator.CalculateCost(orderInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(orderInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	order := common.Order{
		OrderID:    *orderID,
		ExecutorID: *executorID,
		Cost:       cost,
		Payload:    payload,
	}

	err = s.dbProvider.CreateOrder(handlerCtx, &order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) cancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := common.FetchQueryParam(r, common.OrderIDQueryParam)
	handlerCtx := r.Context()

	if orderID == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := s.dbProvider.CancelOrder(handlerCtx, *orderID)
	if err != nil {
		if errors.Is(err, db.ErrNoSuchRowToUpdate) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Write(payload)
}

func (s *Server) setupRoutes() {
	s.mux = chi.NewRouter()

	common.SetupMiddlewares(s.mux)
	// -- add request_id middleware and logs
	// -- think about copypaste in executor
	s.mux.Post("/v1/assign_order", s.assignOrderHandler)
	s.mux.Post("/v1/cancel_order", s.cancelOrderHandler)
}

func NewServer(
	cfg *config.HTTPServerConfig,
	csProvider common.CacheServiceProvider,
	dbProvider common.DBProvider,
	costCalculator common.CostCalculator,
) (*Server, error) {
	server := &Server{
		cfg:            cfg,
		csProvider:     csProvider,
		costCalculator: costCalculator,
		dbProvider:     dbProvider,
	}

	server.setupRoutes()

	return server, nil
}

func (s *Server) Run() error {
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port), s.mux)
}
