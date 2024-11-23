package cacheservice

import (
	"context"
	"errors"
	"fmt"
	"malomopa/internal/config"
	"reflect"
	"testing"
	"time"
)

type mockDataSourceProvider struct {
	name2Getter map[string]DataSourceGetter
}

type mockGetter struct {
	t            *testing.T
	name         string
	expectedDeps map[string]any
	ok           bool
	res          any
	err          error
}

func newMockDataSourceProvider() *mockDataSourceProvider {
	return &mockDataSourceProvider{
		name2Getter: make(map[string]DataSourceGetter),
	}
}

func (m *mockDataSourceProvider) addGet(getter *mockGetter) {
	m.name2Getter[getter.name] = getter
}

func (m *mockDataSourceProvider) GetGet(key string) (DataSourceGetter, error) {
	getter, ok := m.name2Getter[key]
	if !ok {
		return nil, fmt.Errorf("no getter for key %q", key)
	}
	return getter, nil
}

func (m *mockDataSourceProvider) addOkGetter(t *testing.T, name string, res any, expected map[string]any) {
	m.addGet(&mockGetter{
		t:            t,
		name:         name,
		ok:           true,
		res:          res,
		expectedDeps: expected,
	})
}

func (m *mockDataSourceProvider) addFailGetter(t *testing.T, name string, err error, expected map[string]any) {
	m.addGet(&mockGetter{
		t:            t,
		name:         name,
		ok:           false,
		err:          err,
		expectedDeps: expected,
	})
}

func (g *mockGetter) Get(r *DataSourcesRequest, dctx *DataSourceContext) (any, error) {
	if g.expectedDeps != nil && !reflect.DeepEqual(dctx.Deps, g.expectedDeps) {
		g.t.Errorf("Expected %v but got %v deps for %s", g.expectedDeps, dctx.Deps, g.name)
	}
	if g.ok {
		return g.res, nil
	}

	return nil, g.err
}

func newSimpleDataSourceConfig(name string, deps []string) *config.DataSourceConfig {
	return &config.DataSourceConfig{
		Name: name,
		Deps: deps,
	}
}

func newTestCacheServiceConfig() *config.CacheServiceConfig {
	return &config.CacheServiceConfig{
		DataSources: []*config.DataSourceConfig{
			newSimpleDataSourceConfig("A", nil),
			newSimpleDataSourceConfig("B", []string{"A"}),
			newSimpleDataSourceConfig("C", []string{"A"}),
			newSimpleDataSourceConfig("D", []string{"B"}),
			newSimpleDataSourceConfig("E", []string{"B", "C"}),
		},
		GlobalTimeout:  time.Second,
		MaxParallelism: -1,
	}
}

//     A
//    / \
//   B   C
//  / \ /
// D   E

func TestFetchOk(t *testing.T) {
	cfg := newTestCacheServiceConfig()
	provider := newMockDataSourceProvider()
	provider.addOkGetter(t, "A", "Ares", map[string]any{})
	provider.addOkGetter(t, "B", "Bres", map[string]any{"A": "Ares"})
	provider.addOkGetter(t, "C", "Cres", map[string]any{"A": "Ares"})
	provider.addOkGetter(t, "D", "Dres", map[string]any{"B": "Bres"})
	provider.addOkGetter(t, "E", "Eres", map[string]any{"B": "Bres", "C": "Cres"})

	cacheService, err := NewCacheService(cfg, provider)
	if err != nil {
		t.Fatalf("expected no err, got: %v", err)
	}

	fetched, err := cacheService.GetOrderInfo(context.Background(), "kek", "lol")
	if err != nil {
		t.Errorf("expected no err, got: %v", err)
	}

	expected := map[string]string{
		"A": "Ares",
		"B": "Bres",
		"C": "Cres",
		"D": "Dres",
		"E": "Eres",
	}

	if !compareJSONs(fetched, expected) {
		t.Errorf("expected: %v, got: %v", expected, fetched)
	}
}

//     A
//    / \
//   B*  C
//  / \ /
// D   E

func TestFetchingFailures(t *testing.T) {
	cfg := newTestCacheServiceConfig()
	provider := newMockDataSourceProvider()
	provider.addOkGetter(t, "A", "Ares", map[string]any{})
	provider.addFailGetter(t, "B", errors.New("Berr"), map[string]any{"A": "Ares"})
	provider.addOkGetter(t, "C", "Cres", map[string]any{"A": "Ares"})
	provider.addOkGetter(t, "D", "Dres", map[string]any{"B": "Bres"})
	provider.addOkGetter(t, "E", "Eres", map[string]any{"B": "Bres", "C": "Cres"})

	cacheService, err := NewCacheService(cfg, provider)
	if err != nil {
		t.Fatalf("expected no err, got: %v", err)
	}

	_, err = cacheService.GetOrderInfo(context.Background(), "kek", "lol")

	if err == nil {
		t.Errorf("expected error but got: nil")
	}
}

// func TestFetchTimeout(t *testing.T) {
// 	cfg := &config.CacheServiceConfig{}
// 	cacheService, _ := MakeCacheService(cfg)
// 	restore := resetFetchers()
// 	defer restore()

// 	get := newOkGet(t, "A", "Ares", map[fetcherID]any{})
// 	get.RegisterFull(nil, nil, time.Second*1)

// 	get.Delay = time.Second / 2
// 	_, err := cacheService.GetOrderInfo(context.Background(), "kek", "lol")
// 	if err != nil {
// 		t.Errorf("expected: nil, got: %s", err.Error())
// 	}

// 	get.Delay = time.Second * 2
// 	_, err = cacheService.GetOrderInfo(context.Background(), "kek", "lol")
// 	if err == nil {
// 		t.Errorf("expected: not nil, got: nil")
// 	}
// }

// func TestCache(t *testing.T) {
// 	cfg := &config.CacheServiceConfig{}
// 	cacheService, _ := MakeCacheService(cfg)
// 	restore := resetFetchers()
// 	defer restore()

// 	cacheCfg := &fetcherCacheConfig{
// 		maxSize: 10,
// 		ttl:     time.Duration(time.Second * 1),
// 	}

// 	get := newOkGet(t, "A", "Ares", map[fetcherID]any{})
// 	timeout := time.Second * 1
// 	f := get.RegisterFull(nil, cacheCfg, timeout)

// 	// Fill cache
// 	_, err := cacheService.GetOrderInfo(context.Background(), "kek", "lol")
// 	if err != nil {
// 		t.Errorf("expected: nil, got: %s", err.Error())
// 	}

// 	if f.GetCache.cache.GetSize() != 1 {
// 		t.Errorf("expected cache size to be 1, got %d", f.GetCache.cache.GetSize())
// 	}

// 	// Set large delay and expect cache hit
// 	get.Delay = timeout * 2
// 	_, err = cacheService.GetOrderInfo(context.Background(), "kek", "lol")
// 	if err != nil {
// 		t.Errorf("expected: nil, got: %s", err.Error())
// 	}

// 	// Wait cache expiration, expect timeout error
// 	time.Sleep(cacheCfg.ttl * 2)
// 	_, err = cacheService.GetOrderInfo(context.Background(), "kek", "lol")
// 	if err == nil {
// 		t.Errorf("expected: not nil, got nil")
// 	}
// }
