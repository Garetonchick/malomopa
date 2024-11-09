package common

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func FetchQueryParam(r *http.Request, queryParamName string) *string {
	queryParams := r.URL.Query()

	queryParam := queryParams.Get(queryParamName)
	if queryParams.Has(queryParamName) {
		return &queryParam
	}
	return nil
}

func SetupMiddlewares(mux *chi.Mux) {
	mux.Use(middleware.RequestID)
	mux.Use(middleware.Logger)
}
