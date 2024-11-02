package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Garetonchick/malomopa/cache-service/internal/config"
	"github.com/Garetonchick/malomopa/cache-service/internal/fetch"
)

func getAssignOrderInfoHandler(w http.ResponseWriter, r *http.Request) {
	args := struct {
		OrderID    string `json:"order_id"`
		ExecutorID string `json:"executor_id"`
	}{}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(b, &args)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res := fetch.TryAll(context.Background(), args.OrderID, args.ExecutorID)
	b, err = json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /get_assign_order_info", getAssignOrderInfoHandler)

	http.ListenAndServe(config.Host+":"+config.Port, mux)
}
