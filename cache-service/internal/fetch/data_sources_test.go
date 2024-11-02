package fetch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Garetonchick/malomopa/cache-service/internal/config"
	"github.com/Garetonchick/malomopa/cache-service/internal/fetch"
)

type endpointMocks struct {
	t *testing.T
}

var orderID = "aboba"
var executorID = "malomopa"

var expected map[string]any = map[string]any{
	"general_order_info": fetch.GeneralOrderInfo{
		ID:             orderID,
		UserID:         "gareton",
		ZoneID:         "infra",
		BaseCoinAmount: 4.4,
	},
	"zone_info": fetch.ZoneInfo{
		ID:          "infra",
		CoinCoeff:   42.2,
		DisplayName: "INFRAAAAAAAAAA",
	},
	"executor_profile": fetch.ExecutorProfile{
		ID:     executorID,
		Tags:   []string{"goshandr", "mop", "18+"},
		Rating: 100.100,
	},
	"configs": map[string]any{
		"param1": 4,
		"param2": "hhhhhhh",
	},
	"toll_roads_info": fetch.TollRoadsInfo{
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
	reqData := struct {
		OrderID string `json:"id"`
	}{}
	e.readJSON(r, &reqData)
	if reqData.OrderID != orderID {
		e.t.Errorf("wrong order id, expected: %s, got: %s", orderID, reqData.OrderID)
	}
	e.respondWithJSON(w, expected["general_order_info"])
}

func (e endpointMocks) getZoneInfoEndpoint(w http.ResponseWriter, r *http.Request) {
	reqData := struct {
		ZoneID string `json:"id"`
	}{}
	e.readJSON(r, &reqData)
	expectedZoneInfoID := expected["zone_info"].(fetch.ZoneInfo).ID
	if reqData.ZoneID != expectedZoneInfoID {
		e.t.Errorf(
			"wrong zone id, expected: %s, got: %s", expectedZoneInfoID, reqData.ZoneID,
		)
	}
	e.respondWithJSON(w, expected["zone_info"])
}

func (e endpointMocks) getExecutorProfileEndpoint(
	w http.ResponseWriter, r *http.Request,
) {
	reqData := struct {
		ExecutorID string `json:"id"`
	}{}
	e.readJSON(r, &reqData)
	if reqData.ExecutorID != executorID {
		e.t.Errorf(
			"wrong executor id, expected: %s, got: %s", executorID, reqData.ExecutorID,
		)
	}
	e.respondWithJSON(w, expected["executor_profile"])
}

func (e endpointMocks) getConfigsEndpoint(w http.ResponseWriter, r *http.Request) {
	e.respondWithJSON(w, expected["configs"])
}

func (e endpointMocks) getTollRoadsEndpoint(w http.ResponseWriter, r *http.Request) {
	reqData := struct {
		ZoneDisplayName string `json:"zone_display_name"`
	}{}
	e.readJSON(r, &reqData)

	eZoneDisplayName := expected["zone_info"].(fetch.ZoneInfo).DisplayName

	if reqData.ZoneDisplayName != eZoneDisplayName {
		e.t.Errorf(
			"wrong zone display name, expected: %s, got: %s",
			eZoneDisplayName,
			reqData.ZoneDisplayName,
		)
	}
	e.respondWithJSON(w, expected["toll_roads_info"])
}

func compareJSONs(a, b any) bool {
	by1, err := json.Marshal(a)
	if err != nil {
		panic("mop")
	}
	by2, err := json.Marshal(b)
	if err != nil {
		panic("mop")
	}

	return bytes.Equal(by1, by2)
}

func TestDataSources(t *testing.T) {
	getGeneralOrderInfoEndpoint := "/order-info"
	getZoneInfoEndpoint := "/zone-info"
	getExecutorProfileEndpoint := "/executor-profile"
	getConfigsEndpoint := "/configs"
	getTollRoadsInfoEndpoint := "/toll-roads"
	m := endpointMocks{t: t}

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+getGeneralOrderInfoEndpoint, m.getGeneralOrderInfoEndpoint)
	mux.HandleFunc("GET "+getZoneInfoEndpoint, m.getZoneInfoEndpoint)
	mux.HandleFunc("GET "+getExecutorProfileEndpoint, m.getExecutorProfileEndpoint)
	mux.HandleFunc("GET "+getConfigsEndpoint, m.getConfigsEndpoint)
	mux.HandleFunc("GET "+getTollRoadsInfoEndpoint, m.getTollRoadsEndpoint)

	svr := httptest.NewServer(mux)

	config.GetGeneralOrderInfoEndpoint = svr.URL + getGeneralOrderInfoEndpoint
	config.GetZoneInfoEndpoint = svr.URL + getZoneInfoEndpoint
	config.GetExecutorProfileEndpoint = svr.URL + getExecutorProfileEndpoint
	config.GetConfigsEndpoint = svr.URL + getConfigsEndpoint
	config.GetTollRoadsInfoEndpoint = svr.URL + getTollRoadsInfoEndpoint

	fetched := fetch.TryAll(context.Background(), orderID, executorID)

	if !compareJSONs(fetched, expected) {
		t.Errorf("expected: %v, got: %v", expected, fetched)
	}
}
