package cacheservice

import (
	"context"
	"encoding/json"
	"io"
	"malomopa/internal/common"
	"net/http"
	"net/http/httptest"
	"testing"
)

type endpointMocks struct {
	t *testing.T
}

var orderID = "aboba"
var executorID = "malomopa"

var expected map[string]any = map[string]any{
	common.Keys.GeneralOrderInfo: common.GeneralOrderInfo{
		ID:             orderID,
		UserID:         "gareton",
		ZoneID:         "infra",
		BaseCoinAmount: 4.4,
	},
	common.Keys.ZoneInfo: common.ZoneInfo{
		ID:          "infra",
		CoinCoeff:   42.2,
		DisplayName: "INFRAAAAAAAAAA",
	},
	common.Keys.ExecutorProfile: common.ExecutorProfile{
		ID:     executorID,
		Tags:   []string{"goshandr", "mop", "18+"},
		Rating: 100.100,
	},
	common.Keys.AssignOrderConfigs: common.AssignOrderConfigs{
		CoinCoeffCfg: &common.CoinCoeffConfig{
			Max: 1.8,
		},
	},
	common.Keys.TollRoadsInfo: common.TollRoadsInfo{
		BonusAmount: 300.0,
	},
}

func (e *endpointMocks) respondWithJSON(w http.ResponseWriter, data any) {
	b, err := json.Marshal(data)
	if err != nil {
		e.t.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		e.t.Fatal(err)
	}
}

func (e *endpointMocks) readJSON(r *http.Request, v any) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		e.t.Fatal(err)
	}
	err = json.Unmarshal(b, v)
	if err != nil {
		e.t.Fatal(err)
	}
}

func (e endpointMocks) getGeneralOrderInfoEndpoint(
	w http.ResponseWriter, r *http.Request,
) {
	reqData := generalOrderInfoRequest{}
	e.readJSON(r, &reqData)
	if reqData.OrderID != orderID {
		e.t.Errorf("wrong order id, expected: %s, got: %s", orderID, reqData.OrderID)
	}
	e.respondWithJSON(w, expected[common.Keys.GeneralOrderInfo])
}

func (e endpointMocks) getZoneInfoEndpoint(w http.ResponseWriter, r *http.Request) {
	reqData := zoneInfoRequest{}
	e.readJSON(r, &reqData)
	expectedZoneInfoID := expected[common.Keys.ZoneInfo].(common.ZoneInfo).ID
	if reqData.ZoneID != expectedZoneInfoID {
		e.t.Errorf(
			"wrong zone id, expected: %s, got: %s", expectedZoneInfoID, reqData.ZoneID,
		)
	}
	e.respondWithJSON(w, expected[common.Keys.ZoneInfo])
}

func (e endpointMocks) getExecutorProfileEndpoint(
	w http.ResponseWriter, r *http.Request,
) {
	reqData := executorProfileRequest{}
	e.readJSON(r, &reqData)
	if reqData.ExecutorID != executorID {
		e.t.Errorf(
			"wrong executor id, expected: %s, got: %s", executorID, reqData.ExecutorID,
		)
	}
	e.respondWithJSON(w, expected[common.Keys.ExecutorProfile])
}

func (e endpointMocks) getAssignOrderConfigsEndpoint(w http.ResponseWriter, r *http.Request) {
	e.respondWithJSON(w, expected[common.Keys.AssignOrderConfigs])
}

func (e endpointMocks) getTollRoadsEndpoint(w http.ResponseWriter, r *http.Request) {
	reqData := tollRoadsInfoRequest{}
	e.readJSON(r, &reqData)

	eZoneDisplayName := expected[common.Keys.ZoneInfo].(common.ZoneInfo).DisplayName

	if reqData.ZoneDisplayName != eZoneDisplayName {
		e.t.Errorf(
			"wrong zone display name, expected: %s, got: %s",
			eZoneDisplayName,
			reqData.ZoneDisplayName,
		)
	}
	e.respondWithJSON(w, expected[common.Keys.TollRoadsInfo])
}

func TestDataSources(t *testing.T) {
	m := endpointMocks{t: t}
	endpoints := []struct {
		Key      string
		Endpoint string
		Handler  http.HandlerFunc
	}{
		{
			Key:      common.Keys.GeneralOrderInfo,
			Endpoint: "/order-info",
			Handler:  m.getGeneralOrderInfoEndpoint,
		},
		{
			Key:      common.Keys.ZoneInfo,
			Endpoint: "/zone-info",
			Handler:  m.getZoneInfoEndpoint,
		},
		{
			Key:      common.Keys.ExecutorProfile,
			Endpoint: "/executor-profile",
			Handler:  m.getExecutorProfileEndpoint,
		},
		{
			Key:      common.Keys.AssignOrderConfigs,
			Endpoint: "/configs",
			Handler:  m.getAssignOrderConfigsEndpoint,
		},
		{
			Key:      common.Keys.TollRoadsInfo,
			Endpoint: "/toll-roads",
			Handler:  m.getTollRoadsEndpoint,
		},
	}

	mux := http.NewServeMux()

	for _, e := range endpoints {
		mux.HandleFunc("GET "+e.Endpoint, e.Handler)
	}

	svr := httptest.NewServer(mux)

	provider := NewDataSourcesProvider()
	req := DataSourcesRequest{
		OrderID:    orderID,
		ExecutorID: executorID,
	}

	fetched := make(map[string]any)

	get := func(key string, endpoint string) {
		getter, err := provider.GetGet(key)
		if err != nil {
			t.Fatalf("failed to GetGet: %v", err)
		}
		fetched[key], err = getter.Get(&req, &DataSourceContext{
			Ctx:      context.Background(),
			Endpoint: endpoint,
			Cache:    nil,
			Deps:     fetched,
		})
		if err != nil {
			t.Fatalf("expected: nil, got: %v", err)
		}
	}

	for _, e := range endpoints {
		get(e.Key, svr.URL+e.Endpoint)
	}

	if !compareJSONs(fetched, expected) {
		t.Errorf("expected: %v, got: %v", expected, fetched)
	}
}
