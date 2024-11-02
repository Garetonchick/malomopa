package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Garetonchick/malomopa/cache-service/internal/config"
)

var getGeneralOrderInfoF *fetcher
var getZoneInfoF *fetcher
var getExecutorProfileF *fetcher
var getConfigsF *fetcher
var getTollRoadsInfoF *fetcher

type GeneralOrderInfo struct {
	ID             string  `json:"id"`
	UserID         string  `json:"user_id"`
	ZoneID         string  `json:"zone_id"`
	BaseCoinAmount float32 `json:"base_coin_amount"`
}

type ZoneInfo struct {
	ID          string  `json:"id"`
	CoinCoeff   float32 `json:"coin_coeff"`
	DisplayName string  `json:"display_name"`
}

type ExecutorProfile struct {
	ID     string   `json:"id"`
	Tags   []string `json:"tags"`
	Rating float32  `json:"rating"`
}

type TollRoadsInfo struct {
	BonusAmount float32 `json:"bonus_amount"`
}

func doJSONRequest(ctx context.Context, data any, endpoint string, v any) error {
	var err error
	var b []byte
	if data != nil {
		b, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	var req *http.Request

	if data != nil {
		req, err = http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			endpoint,
			bytes.NewReader(b),
		)
	} else {
		req, err = http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			endpoint,
			nil,
		)
	}

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

	if err = json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil

}

// TODO: Add timeout
func getGeneralOrderInfo(c *call, deps map[fetcherID]any) (any, error) {
	var info GeneralOrderInfo
	err := doJSONRequest(
		c.Ctx,
		map[string]string{"id": c.OrderID},
		config.GetGeneralOrderInfoEndpoint,
		&info,
	)
	return info, err
}

// TODO: Add cache and timeout
func getZoneInfo(c *call, deps map[fetcherID]any) (any, error) {
	orderInfo := deps[getGeneralOrderInfoF.ID].(GeneralOrderInfo)

	var info ZoneInfo
	err := doJSONRequest(
		c.Ctx,
		map[string]string{"id": orderInfo.ZoneID},
		config.GetZoneInfoEndpoint,
		&info,
	)
	return info, err
}

// TODO: Add timeout
func getExecutorProfile(c *call, deps map[fetcherID]any) (any, error) {
	var profile ExecutorProfile
	err := doJSONRequest(
		c.Ctx,
		map[string]string{"id": c.ExecutorID},
		config.GetExecutorProfileEndpoint,
		&profile,
	)
	return profile, err
}

// TODO: Add cache and timeout
func getConfigs(c *call, deps map[fetcherID]any) (any, error) {
	var configs map[string]any
	err := doJSONRequest(
		c.Ctx,
		nil,
		config.GetConfigsEndpoint,
		&configs,
	)
	return configs, err
}

// TODO: Add timeout
func getTollRoadsInfo(c *call, deps map[fetcherID]any) (any, error) {
	zoneInfo := deps[getZoneInfoF.ID].(ZoneInfo)

	var info TollRoadsInfo
	err := doJSONRequest(
		c.Ctx,
		map[string]string{"zone_display_name": zoneInfo.DisplayName},
		config.GetTollRoadsInfoEndpoint,
		&info,
	)
	return info, err
}

func init() {
	// disable unused warnings
	_ = getExecutorProfileF
	_ = getConfigsF
	_ = getTollRoadsInfoF

	getGeneralOrderInfoF = registerFetcher(
		getGeneralOrderInfo, "general_order_info", nil,
	)
	getZoneInfoF = registerFetcher(
		getZoneInfo, "zone_info", []*fetcher{getGeneralOrderInfoF},
	)
	getExecutorProfileF = registerFetcher(
		getExecutorProfile, "executor_profile", nil,
	)
	getConfigsF = registerFetcher(
		getConfigs, "configs", nil,
	)
	getTollRoadsInfoF = registerFetcher(
		getTollRoadsInfo, "toll_roads_info", []*fetcher{getZoneInfoF},
	)
}
