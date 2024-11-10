package common

import (
	"context"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type RequestLogger struct {
	*zap.Logger

	globalFields []zap.Field
}

func GetRequestLogger(ctx context.Context, service, methodName string) *RequestLogger {
	if logger, ok := ctx.Value(loggerCtxKey).(*zap.Logger); ok {
		requestID := middleware.GetReqID(ctx)
		return &RequestLogger{
			Logger: logger,
			globalFields: []zap.Field{
				zap.String("request_id", requestID),
				zap.String("service", service),
				zap.String("method", methodName),
			},
		}
	}
	return &RequestLogger{}
}

func (l *RequestLogger) IsValid() bool {
	return l.Logger != nil
}

func (l *RequestLogger) AddGlobalFields(fields ...zap.Field) {
	if l.Logger == nil {
		return
	}

	l.globalFields = append(l.globalFields, fields...)
}

func (l *RequestLogger) Error(msg string, fields ...zap.Field) {
	if l.Logger == nil {
		return
	}

	l.Logger.Error(msg, append(fields, l.globalFields...)...)
}

func (l *RequestLogger) Info(msg string, fields ...zap.Field) {
	if l.Logger == nil {
		return
	}

	l.Logger.Info(msg, append(fields, l.globalFields...)...)
}
