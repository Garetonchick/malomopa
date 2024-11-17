package common

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func FetchQueryParam(r *http.Request, queryParamName string) *string {
	queryParams := r.URL.Query()

	queryParam := queryParams.Get(queryParamName)
	if queryParams.Has(queryParamName) {
		return &queryParam
	}
	return nil
}

func TerminateWithErr(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

type loggerCtxKeyType uint8
type startTsCtxKeyType uint8

var loggerCtxKey loggerCtxKeyType
var startTsCtxKey startTsCtxKeyType

func SetupMiddlewares(mux *chi.Mux, logger *zap.Logger) {
	// order is important
	mux.Use(
		middleware.WithValue(loggerCtxKey, logger),
		middleware.RequestID,
		logIncomingRequest,
		logOutgoingResponse,
	)
}

func logIncomingRequest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := GetRequestLogger(ctx, "", "")
		logger.Info("query started")

		ctx = context.WithValue(ctx, startTsCtxKey, time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func logOutgoingResponse(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		ctx := r.Context()
		logger := GetRequestLogger(ctx, "", "")

		var elapsed time.Duration
		if startTs, ok := ctx.Value(startTsCtxKey).(time.Time); ok {
			elapsed = time.Since(startTs)
		}

		logger.Info("query finished",
			zap.Duration("elapsed", elapsed),
		)
	}
	return http.HandlerFunc(fn)
}
