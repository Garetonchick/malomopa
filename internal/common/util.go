package common

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func DoJSONRequest(ctx context.Context, endpoint string, data any, out any) error {
	var err error
	var b []byte
	if data != nil {
		b, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	reqBody := bytes.NewReader([]byte{})
	if data != nil {
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		reqBody,
	)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, out)
}

func Camel2Snake(camel string) string {
	snake := strings.Builder{}
	snake.Grow(len(camel) + 4)
	prevUpper := true
	for _, r := range camel {
		upper := unicode.IsUpper(r)
		if upper && !prevUpper {
			snake.WriteRune('_')
		}
		snake.WriteRune(unicode.ToLower(r))
		prevUpper = upper
	}
	return snake.String()
}

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

func ReadJSONFromFile[T any](path string) (T, error) {
	file, err := os.Open(path)
	var v T
	if err != nil {
		return v, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return v, err
	}

	err = json.Unmarshal(data, &v)
	return v, err
}

func WriteJSONToFile(path string, value any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	return encoder.Encode(value)
}
