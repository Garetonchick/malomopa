package common

import (
	"context"
	"net/http"

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

type loggerCtxKeyType uint8

var loggerCtxKey loggerCtxKeyType

func SetupMiddlewares(mux *chi.Mux, logger *zap.Logger) {
	mux.Use(middleware.RequestID, middleware.WithValue(loggerCtxKey, logger))
}

func GetRequestLogger(ctx context.Context) *RequestLogger {
	if logger, ok := ctx.Value(loggerCtxKey).(*zap.Logger); ok {
		requestID := middleware.GetReqID(ctx)
		return &RequestLogger{
			Logger:       logger,
			globalFields: []zap.Field{zap.String("request_id", requestID)},
		}
	}
	return &RequestLogger{}
}

type RequestLogger struct {
	*zap.Logger

	globalFields []zap.Field
}

func (l *RequestLogger) IsValid() bool {
	return l.Logger != nil
}

func (l *RequestLogger) AddGlobalFields(fields ...zap.Field) {
	l.globalFields = append(l.globalFields, fields...)
}

func (l *RequestLogger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, append(fields, l.globalFields...)...)
}

func (l *RequestLogger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, append(fields, l.globalFields...)...)
}
