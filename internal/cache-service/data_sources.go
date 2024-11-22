package cacheservice

import (
	"context"
	"fmt"
	"malomopa/internal/common"
	"reflect"
	"strings"
)

// When adding new data source make sure that
// method name is in the form of "get<KEY>"
// where KEY is this data source's key in Camel case
// e. g. for "general_order_info" data source the
// correct method name is "getGeneralOrderInfo".
// In case the name is wrong, it won't be registered
// properly via reflection.
type dataSourcesProvider struct {
	dataSourceKey2Get map[string]dataSourceGet
}

type dataSourcesRequester struct {
	provider   *dataSourcesProvider
	orderID    string
	executorID string
}

type dataSourceContext struct {
	ctx      context.Context
	cache    Cache
	endpoint string
	deps     map[string]any
}

type dataSourceGet = func(*dataSourcesRequester, *dataSourceContext) (any, error)

type generalOrderInfoRequest struct {
	OrderID string `json:"id"`
}

type zoneInfoRequest struct {
	ZoneID string `json:"id"`
}

type executorProfileRequest struct {
	ExecutorID string `json:"id"`
}

type tollRoadsInfoRequest struct {
	ZoneDisplayName string `json:"zone_display_name"`
}

func newDataSourcesProvider() *dataSourcesProvider {
	// eto infra
	defer func() {
		message := "Detected data sources configuration error.\n" +
			"Recheck keys and names of data source methods.\n" +
			"Additional error context: %v"

		if err := recover(); err != nil {
			panic(fmt.Sprintf(message, err))
		}
	}()
	provider := dataSourcesProvider{
		dataSourceKey2Get: make(map[string]dataSourceGet),
	}

	methodName2KeyName := func(mname string) string {
		return common.Camel2Snake(strings.TrimPrefix(mname, "get"))
	}

	val := reflect.ValueOf(common.Keys)
	typ := val.Type()
	var keys map[string]bool

	for i := 0; i < typ.NumField(); i++ {
		value := val.Field(i)
		keys[value.String()] = true
	}

	typ = reflect.TypeOf((*dataSourcesRequester)(nil))

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		name := methodName2KeyName(method.Name)

		was, ok := keys[name]
		if !ok {
			panic(fmt.Sprintf(
				"Unknown data source method with name %q. Did you forget to add key for it?",
				name,
			))
		}
		if !was {
			panic(fmt.Sprintf(
				"Found 2 data source methods with the same key name %q.",
				name,
			))
		}
		keys[name] = false

		provider.dataSourceKey2Get[name] = method.Func.Interface().(dataSourceGet)
	}

	if typ.NumMethod() != len(keys) {
		panic(fmt.Sprintf(
			"Keys count (%d) must be equal to the data source methods count (%d).",
			len(keys),
			typ.NumMethod(),
		))
	}

	return &provider
}

func (p *dataSourcesProvider) newRequester(orderID string, executorID string) *dataSourcesRequester {
	return &dataSourcesRequester{
		provider:   p,
		orderID:    orderID,
		executorID: executorID,
	}
}

func (p *dataSourcesProvider) getDataSourceGetByKey(key string) dataSourceGet {
	return p.dataSourceKey2Get[key]
}

func (r *dataSourcesRequester) getGeneralOrderInfo(dctx *dataSourceContext) (any, error) {
	var info common.GeneralOrderInfo
	return genericDataSourceGet(
		dctx,
		r.orderID,
		generalOrderInfoRequest{OrderID: r.orderID},
		&info,
	)
}

func (r *dataSourcesRequester) getZoneInfo(dctx *dataSourceContext) (any, error) {
	orderInfo := getDep[common.GeneralOrderInfo](dctx, common.Keys.GeneralOrderInfo)

	var info common.ZoneInfo
	return genericDataSourceGet(
		dctx,
		orderInfo.ZoneID,
		zoneInfoRequest{ZoneID: orderInfo.ZoneID},
		&info,
	)
}

func (r *dataSourcesRequester) getExecutorProfile(dctx *dataSourceContext) (any, error) {
	var profile common.ExecutorProfile
	return genericDataSourceGet(
		dctx,
		r.executorID,
		executorProfileRequest{ExecutorID: r.executorID},
		&profile,
	)
}

func (r *dataSourcesRequester) getAssignOrderConfigs(dctx *dataSourceContext) (any, error) {
	var configs common.AssignOrderConfigs
	return genericDataSourceGet(
		dctx,
		DefaultCacheKey,
		nil,
		&configs,
	)
}

func (r *dataSourcesRequester) getTollRoadsInfo(dctx *dataSourceContext) (any, error) {
	zoneInfo := getDep[common.ZoneInfo](dctx, common.Keys.ZoneInfo)

	var info common.TollRoadsInfo
	return genericDataSourceGet(
		dctx,
		zoneInfo.DisplayName,
		tollRoadsInfoRequest{ZoneDisplayName: zoneInfo.DisplayName},
		&info,
	)
}

func getDep[T any](d *dataSourceContext, depKey string) T {
	return d.deps[depKey].(T)
}

func genericDataSourceGet[T any](
	dctx *dataSourceContext,
	cacheKey string,
	in any,
	out *T,
) (any, error) {
	return GetFromCacheOrCompute(dctx.cache, cacheKey, func() (any, error) {
		err := common.DoJSONRequest(
			dctx.ctx,
			dctx.endpoint,
			in,
			out,
		)
		return *out, err
	})
}
