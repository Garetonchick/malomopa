package cacheservice

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"malomopa/internal/common"
	"net/http"
)

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
func getGeneralOrderInfo(c *call, endpoint string, deps map[fetcherID]any) (any, error) {
	var info common.GeneralOrderInfo
	err := doJSONRequest(
		c.Ctx,
		map[string]string{"id": c.OrderID},
		endpoint,
		&info,
	)
	return info, err
}

// TODO: Add cache and timeout
func getZoneInfo(c *call, endpoint string, deps map[fetcherID]any) (any, error) {
	orderInfo := deps[getGeneralOrderInfoF.ID].(common.GeneralOrderInfo)

	var info common.ZoneInfo
	err := doJSONRequest(
		c.Ctx,
		map[string]string{"id": orderInfo.ZoneID},
		endpoint,
		&info,
	)
	return info, err
}

// TODO: Add timeout
func getExecutorProfile(c *call, endpoint string, deps map[fetcherID]any) (any, error) {
	var profile common.ExecutorProfile
	err := doJSONRequest(
		c.Ctx,
		map[string]string{"id": c.ExecutorID},
		endpoint,
		&profile,
	)
	return profile, err
}

// TODO: Add cache and timeout
func getConfigs(c *call, endpoint string, deps map[fetcherID]any) (any, error) {
	var configs map[string]any
	err := doJSONRequest(
		c.Ctx,
		nil,
		endpoint,
		&configs,
	)
	return configs, err
}

// TODO: Add timeout
func getTollRoadsInfo(c *call, endpoint string, deps map[fetcherID]any) (any, error) {
	zoneInfo := deps[getZoneInfoF.ID].(common.ZoneInfo)

	var info common.TollRoadsInfo
	err := doJSONRequest(
		c.Ctx,
		map[string]string{"zone_display_name": zoneInfo.DisplayName},
		endpoint,
		&info,
	)
	return info, err
}
