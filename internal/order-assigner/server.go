package assigner

import (
	"encoding/json"
	"fmt"
	"malomopa/internal/common"
	"malomopa/internal/config"
	"net/http"
)

type Server struct {
	cfg *config.OrderAssignerConfig

	mux *http.ServeMux

	dsProvider common.CacheServiceProvider
	dbProvider common.DBProvider
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
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderID := fetchQueryParam(r, common.OrderIDQueryParam)
	executorID := fetchQueryParam(r, common.ExecutorIDQueryParam)

	if orderID == nil || executorID == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order, err := s.dsProvider.GetOrderInfo(*orderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.dbProvider.CreateOrder(order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) cancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderID := fetchQueryParam(r, common.OrderIDQueryParam)

	if orderID == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order, err := s.dbProvider.CancelOrder(*orderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resultBody, err := json.Marshal(order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(resultBody)
}

func (s *Server) setupRoutes() {
	s.mux = http.NewServeMux()

	s.mux.HandleFunc("/v1/assign_order", s.assignOrderHandler)
	s.mux.HandleFunc("/v1/cancel_order", s.cancelOrderHandler)
}

func NewServer(cfg *config.OrderAssignerConfig) (*Server, error) {
	server := &Server{
		cfg: cfg,
	}

	server.setupRoutes()

	return server, nil
}

func (s *Server) Run() error {
	return http.ListenAndServe(":"+fmt.Sprintf("%d", s.cfg.HTTPPort), s.mux)
}
