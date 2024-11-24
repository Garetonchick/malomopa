package cacheservice

import (
	"context"
	"encoding/json"
	"io"
	"malomopa/internal/common"
	"malomopa/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type endpointMocks struct {
	t          *testing.T
	orderID    string
	executorID string
	data       map[string]any
	endpoints  map[string]mockEndpoint
}

type mockEndpoint struct {
	endpoint string
	handler  http.HandlerFunc
}

func newEndpointMocks(t *testing.T, orderID string, executorID string) (*endpointMocks, func()) {
	data := map[string]any{
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

	m := &endpointMocks{
		t:          t,
		orderID:    orderID,
		executorID: executorID,
		data:       data,
	}

	endpoints := map[string]mockEndpoint{
		common.Keys.GeneralOrderInfo: {
			endpoint: "/order-info",
			handler:  m.getGeneralOrderInfoEndpoint,
		},
		common.Keys.ZoneInfo: {
			endpoint: "/zone-info",
			handler:  m.getZoneInfoEndpoint,
		},
		common.Keys.ExecutorProfile: {
			endpoint: "/executor-profile",
			handler:  m.getExecutorProfileEndpoint,
		},
		common.Keys.AssignOrderConfigs: {
			endpoint: "/configs",
			handler:  m.getAssignOrderConfigsEndpoint,
		},
		common.Keys.TollRoadsInfo: {
			endpoint: "/toll-roads",
			handler:  m.getTollRoadsEndpoint,
		},
	}

	mux := http.NewServeMux()

	for _, e := range endpoints {
		mux.HandleFunc("GET "+e.endpoint, e.handler)
	}

	svr := httptest.NewServer(mux)
	for key, e := range endpoints {
		endpoints[key] = mockEndpoint{
			endpoint: svr.URL + e.endpoint,
			handler:  e.handler,
		}
	}

	m.endpoints = endpoints

	return m, svr.Close
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

func (e *endpointMocks) getGeneralOrderInfoEndpoint(
	w http.ResponseWriter, r *http.Request,
) {
	reqData := generalOrderInfoRequest{}
	e.readJSON(r, &reqData)
	if reqData.OrderID != e.orderID {
		e.t.Errorf("wrong order id, expected: %s, got: %s", e.orderID, reqData.OrderID)
	}
	e.respondWithJSON(w, e.data[common.Keys.GeneralOrderInfo])
}

func (e *endpointMocks) getZoneInfoEndpoint(w http.ResponseWriter, r *http.Request) {
	reqData := zoneInfoRequest{}
	e.readJSON(r, &reqData)
	expectedZoneInfoID := e.data[common.Keys.ZoneInfo].(common.ZoneInfo).ID
	if reqData.ZoneID != expectedZoneInfoID {
		e.t.Errorf(
			"wrong zone id, expected: %s, got: %s", expectedZoneInfoID, reqData.ZoneID,
		)
	}
	e.respondWithJSON(w, e.data[common.Keys.ZoneInfo])
}

func (e *endpointMocks) getExecutorProfileEndpoint(
	w http.ResponseWriter, r *http.Request,
) {
	reqData := executorProfileRequest{}
	e.readJSON(r, &reqData)
	if reqData.ExecutorID != e.executorID {
		e.t.Errorf(
			"wrong executor id, expected: %s, got: %s", e.executorID, reqData.ExecutorID,
		)
	}
	e.respondWithJSON(w, e.data[common.Keys.ExecutorProfile])
}

func (e *endpointMocks) getAssignOrderConfigsEndpoint(w http.ResponseWriter, r *http.Request) {
	e.respondWithJSON(w, e.data[common.Keys.AssignOrderConfigs])
}

func (e *endpointMocks) getTollRoadsEndpoint(w http.ResponseWriter, r *http.Request) {
	reqData := tollRoadsInfoRequest{}
	e.readJSON(r, &reqData)

	eZoneDisplayName := e.data[common.Keys.ZoneInfo].(common.ZoneInfo).DisplayName

	if reqData.ZoneDisplayName != eZoneDisplayName {
		e.t.Errorf(
			"wrong zone display name, expected: %s, got: %s",
			eZoneDisplayName,
			reqData.ZoneDisplayName,
		)
	}
	e.respondWithJSON(w, e.data[common.Keys.TollRoadsInfo])
}

func TestDataSources(t *testing.T) {
	provider := NewDataSourcesProvider()
	req := DataSourcesRequest{
		OrderID:    "aboba",
		ExecutorID: "malomopa",
	}

	m, close := newEndpointMocks(t, req.OrderID, req.ExecutorID)
	defer close()

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

	ordering := []string{
		common.Keys.GeneralOrderInfo,
		common.Keys.ZoneInfo,
		common.Keys.ExecutorProfile,
		common.Keys.AssignOrderConfigs,
		common.Keys.TollRoadsInfo,
	}

	for _, key := range ordering {
		get(key, m.endpoints[key].endpoint)
	}

	if !compareJSONs(fetched, m.data) {
		t.Errorf("expected: %v, got: %v", m.data, fetched)
	}
}

func TestDataSourceCacheBasic(t *testing.T) {
	req := DataSourcesRequest{
		OrderID:    "aboba",
		ExecutorID: "malomopa",
	}

	provider := NewDataSourcesProvider()
	getter, err := provider.GetGet(common.Keys.GeneralOrderInfo)
	if err != nil {
		t.Fatalf("failed to GetGet: %v", err)
	}

	cache := NewLRUCache(&config.CacheConfig{
		Name:    LRUCacheName,
		TTL:     time.Millisecond * 100,
		MaxSize: 100,
	})

	getExpectErr := func(endpoint string, req *DataSourcesRequest) {
		_, err := getter.Get(req, &DataSourceContext{
			Ctx:      context.Background(),
			Cache:    cache,
			Endpoint: endpoint,
			Deps:     nil,
		})

		if err == nil {
			t.Fatalf("expected error but didn't get one")
		}
	}

	getExpectNoErr := func(m *endpointMocks, req *DataSourcesRequest) {
		got, err := getter.Get(req, &DataSourceContext{
			Ctx:      context.Background(),
			Cache:    cache,
			Endpoint: m.endpoints[common.Keys.GeneralOrderInfo].endpoint,
			Deps:     nil,
		})

		if err != nil {
			t.Fatalf("expected no error but got: %v", err)
		}

		expected := m.data[common.Keys.GeneralOrderInfo].(common.GeneralOrderInfo)

		if expected != got {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	}

	var invalid_endpoint = "http://0.0.0.0:38921/there-is-nothing-here"

	getExpectErr(invalid_endpoint, &req)

	m, close := newEndpointMocks(t, req.OrderID, req.ExecutorID)
	defer close()

	getExpectNoErr(m, &req)
	close()
	getExpectNoErr(m, &req)
	getExpectNoErr(m, &req)

	req.OrderID = "kek"
	getExpectErr(invalid_endpoint, &req)
}

func TestDataSourceCacheExpiredOrOverflowed(t *testing.T) {
	req := DataSourcesRequest{
		OrderID:    "aboba",
		ExecutorID: "malomopa",
	}

	provider := NewDataSourcesProvider()
	getter, err := provider.GetGet(common.Keys.GeneralOrderInfo)
	if err != nil {
		t.Fatalf("failed to GetGet: %v", err)
	}

	ttl := time.Millisecond * 100

	cache := NewLRUCache(&config.CacheConfig{
		Name:    LRUCacheName,
		TTL:     ttl,
		MaxSize: 1,
	})

	getExpectErr := func(endpoint string, req *DataSourcesRequest) {
		_, err := getter.Get(req, &DataSourceContext{
			Ctx:      context.Background(),
			Cache:    cache,
			Endpoint: endpoint,
			Deps:     nil,
		})

		if err == nil {
			t.Fatalf("expected error but didn't get one")
		}
	}

	getExpectNoErr := func(m *endpointMocks, req *DataSourcesRequest) {
		got, err := getter.Get(req, &DataSourceContext{
			Ctx:      context.Background(),
			Cache:    cache,
			Endpoint: m.endpoints[common.Keys.GeneralOrderInfo].endpoint,
			Deps:     nil,
		})

		if err != nil {
			t.Fatalf("expected no error but got: %v", err)
		}

		expected := m.data[common.Keys.GeneralOrderInfo].(common.GeneralOrderInfo)

		if expected != got {
			t.Errorf("expected: %v, got: %v", expected, got)
		}
	}

	m, close := newEndpointMocks(t, req.OrderID, req.ExecutorID)
	defer close()

	endpoint := m.endpoints[common.Keys.GeneralOrderInfo].endpoint

	getExpectNoErr(m, &req)
	close()
	getExpectNoErr(m, &req)
	getExpectNoErr(m, &req)

	time.Sleep(ttl * 2) // wait till cache expires

	getExpectErr(endpoint, &req)

	// restart server
	m, close = newEndpointMocks(t, req.OrderID, req.ExecutorID)
	defer close()

	endpoint = m.endpoints[common.Keys.GeneralOrderInfo].endpoint

	getExpectNoErr(m, &req)

	var oldOrderID = req.OrderID
	var newOrderID = "kek"

	m.orderID = newOrderID
	req.OrderID = newOrderID

	getExpectNoErr(m, &req) // push data for oldOrderID out of cache
	close()
	getExpectNoErr(m, &req)

	req.OrderID = oldOrderID
	m.orderID = oldOrderID
	getExpectErr(endpoint, &req)
}
