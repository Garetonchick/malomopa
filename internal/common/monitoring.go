package common

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	PreviousRequestCount uint64
	_                    [64]byte
	CurrentRequestCount  uint64
)

func AddRequest() {
	atomic.AddUint64(&CurrentRequestCount, 1)
}

func UpdateRPSWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			atomic.StoreUint64(&PreviousRequestCount, atomic.LoadUint64(&CurrentRequestCount))
			atomic.StoreUint64(&CurrentRequestCount, 0)
		case <-ctx.Done():
			return
		}
	}
}

type MonitoringServer struct {
}

func (s *MonitoringServer) SetupRoutes(mux *chi.Mux) {
	mux.Get("/ping", s.pingHandler)
	mux.Get("/rps", s.rpsHandler)
}

func (s *MonitoringServer) pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("ping\n"))
}

func (s *MonitoringServer) rpsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(strconv.FormatUint(atomic.LoadUint64(&PreviousRequestCount), 10) + "\n"))
}
