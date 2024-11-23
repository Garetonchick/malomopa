package cacheservice

import (
	"context"
	"fmt"
	"malomopa/internal/common"
	"reflect"
	"strings"
	"time"

	"go.uber.org/zap"
)

type DataSourcesProvider interface {
	GetGet(key string) (DataSourceGetter, error)
}

type DataSourceGetter interface {
	Get(*DataSourcesRequest, *DataSourceContext) (any, error)
}

type DataSourcesRequest struct {
	Logger     *common.RequestLogger
	OrderID    string
	ExecutorID string
}

type DataSourceContext struct {
	Ctx      context.Context
	Cache    Cache
	Endpoint string
	Deps     map[string]any
}

// When adding new data source make sure that
// method name is in the form of "Get<KEY>"
// where KEY is this data source's key in Camel case
// e. g. for "general_order_info" data source the
// correct method name is "GetGeneralOrderInfo".
// In case the name is wrong, it won't be registered
// properly via reflection.
type dataSourcesProviderImpl struct {
	dataSourceKey2Getter map[string]DataSourceGetter
}

type rawDataSourceGet = func(*dataGetters, *DataSourcesRequest, *DataSourceContext) (any, error)

type getterImpl struct {
	get rawDataSourceGet
}

type dataGetters struct{}

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

func NewDataSourcesProvider() DataSourcesProvider {
	// eto infra
	defer func() {
		message := "Detected data sources configuration error.\n" +
			"Recheck keys and names of data source methods.\n" +
			"Additional error context: %v"

		if err := recover(); err != nil {
			panic(fmt.Sprintf(message, err))
		}
	}()
	provider := dataSourcesProviderImpl{
		dataSourceKey2Getter: make(map[string]DataSourceGetter),
	}

	methodName2KeyName := func(mname string) string {
		return common.Camel2Snake(strings.TrimPrefix(mname, "Get"))
	}

	val := reflect.ValueOf(common.Keys)
	typ := val.Type()
	keys := make(map[string]bool)

	for i := 0; i < typ.NumField(); i++ {
		value := val.Field(i)
		keys[value.String()] = true
	}

	typ = reflect.TypeOf((*dataGetters)(nil))

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		name := methodName2KeyName(method.Name)

		was, ok := keys[name]
		if !ok {
			panic(fmt.Sprintf(
				"Unknown data source method with name %q and respective key name %q. "+
					"Did you forget to add key for it?",
				method.Name, name,
			))
		}
		if !was {
			panic(fmt.Sprintf(
				"Found 2 data source methods with the same key name %q.",
				name,
			))
		}
		keys[name] = false

		rawGet := method.Func.Interface().(rawDataSourceGet)
		provider.dataSourceKey2Getter[name] = &getterImpl{rawGet}
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

func (p *dataSourcesProviderImpl) GetGet(key string) (DataSourceGetter, error) {
	getter, ok := p.dataSourceKey2Getter[key]
	if !ok {
		return nil, fmt.Errorf("there is no getter for key %q", key)
	}
	return getter, nil
}

func (*dataGetters) GetGeneralOrderInfo(r *DataSourcesRequest, dctx *DataSourceContext) (any, error) {
	var info common.GeneralOrderInfo
	return genericDataSourceGet(
		common.Keys.GeneralOrderInfo,
		r,
		dctx,
		r.OrderID,
		generalOrderInfoRequest{OrderID: r.OrderID},
		&info,
	)
}

func (*dataGetters) GetZoneInfo(r *DataSourcesRequest, dctx *DataSourceContext) (any, error) {
	orderInfo := getDep[common.GeneralOrderInfo](dctx, common.Keys.GeneralOrderInfo)

	var info common.ZoneInfo
	return genericDataSourceGet(
		common.Keys.ZoneInfo,
		r,
		dctx,
		orderInfo.ZoneID,
		zoneInfoRequest{ZoneID: orderInfo.ZoneID},
		&info,
	)
}

func (*dataGetters) GetExecutorProfile(r *DataSourcesRequest, dctx *DataSourceContext) (any, error) {
	var profile common.ExecutorProfile
	return genericDataSourceGet(
		common.Keys.ExecutorProfile,
		r,
		dctx,
		r.ExecutorID,
		executorProfileRequest{ExecutorID: r.ExecutorID},
		&profile,
	)
}

func (*dataGetters) GetAssignOrderConfigs(r *DataSourcesRequest, dctx *DataSourceContext) (any, error) {
	var configs common.AssignOrderConfigs
	return genericDataSourceGet(
		common.Keys.AssignOrderConfigs,
		r,
		dctx,
		DefaultCacheKey,
		nil,
		&configs,
	)
}

func (*dataGetters) GetTollRoadsInfo(r *DataSourcesRequest, dctx *DataSourceContext) (any, error) {
	zoneInfo := getDep[common.ZoneInfo](dctx, common.Keys.ZoneInfo)

	var info common.TollRoadsInfo
	return genericDataSourceGet(
		common.Keys.TollRoadsInfo,
		r,
		dctx,
		zoneInfo.DisplayName,
		tollRoadsInfoRequest{ZoneDisplayName: zoneInfo.DisplayName},
		&info,
	)
}

func getDep[T any](d *DataSourceContext, depKey string) T {
	return d.Deps[depKey].(T)
}

func genericDataSourceGet[T any](
	name string,
	r *DataSourcesRequest,
	dctx *DataSourceContext,
	cacheKey string,
	in any,
	out *T,
) (any, error) {
	fromCache := true

	res, err := GetFromCacheOrCompute(dctx.Cache, cacheKey, func() (any, error) {
		startTime := time.Now()
		fromCache = false
		if r.Logger != nil {
			r.Logger.Info(
				"starting fetch from data source",
				zap.String("data_source", name),
			)
		}

		err := common.DoJSONRequest(
			dctx.Ctx,
			dctx.Endpoint,
			in,
			out,
		)

		if err != nil && r.Logger != nil {
			r.Logger.Info(
				"failed fetch from data source",
				zap.String("data_source", name),
				zap.String("error", err.Error()),
			)
		} else if r.Logger != nil {
			r.Logger.Info(
				"fetched data source",
				zap.String("data_source", name),
				zap.Duration("fetch_time", time.Since(startTime)),
			)
		}
		return *out, err
	})

	if fromCache {
		if err != nil && r.Logger != nil {
			r.Logger.Info(
				"failed to retrieve data source from cache",
				zap.String("data_source", name),
				zap.String("cache_key", cacheKey),
				zap.String("error", err.Error()),
			)
		} else if r.Logger != nil {
			r.Logger.Info(
				"retrieved data from cache",
				zap.String("data_source", name),
				zap.String("cache_key", cacheKey),
			)
		}
	}
	return res, err
}

func (g *getterImpl) Get(r *DataSourcesRequest, dctx *DataSourceContext) (any, error) {
	return g.get(nil, r, dctx)
}
