package assigner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"malomopa/internal/common"
	"malomopa/internal/config"
	"malomopa/internal/db"
	"net/http"
)

type Server struct {
	cfg *config.HTTPServerConfig

	mux *http.ServeMux

	csProvider     common.CacheServiceProvider
	costCalculator common.CostCalculator
	dbProvider     common.DBProvider
}

func fetchQueryParam(r *http.Request, queryParamName string) *string {
	queryParams := r.URL.Query()

	queryParam := queryParams.Get(queryParamName)
	if queryParams.Has(queryParamName) {
		return &queryParam
	}
	return nil
}

func (s *Server) assignOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := fetchQueryParam(r, common.OrderIDQueryParam)
	executorID := fetchQueryParam(r, common.ExecutorIDQueryParam)

	if orderID == nil || executorID == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderInfo, err := s.csProvider.GetOrderInfo(context.TODO(), *orderID, *executorID)
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

	err = s.dbProvider.CreateOrder(&order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) cancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := fetchQueryParam(r, common.OrderIDQueryParam)

	if orderID == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := s.dbProvider.CancelOrder(*orderID)
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
	s.mux = http.NewServeMux()

	s.mux.HandleFunc("POST /v1/assign_order", s.assignOrderHandler)
	s.mux.HandleFunc("POST /v1/cancel_order", s.cancelOrderHandler)
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
