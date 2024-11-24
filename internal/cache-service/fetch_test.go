package cacheservice

import (
	"context"
	"errors"
	"fmt"
	"malomopa/internal/common"
	"malomopa/internal/config"
	"reflect"
	"testing"
	"time"
)

type mockDataSourceProvider struct {
	name2Getter map[string]DataSourceGetter
	delay       time.Duration
}

type mockGetter struct {
	t            *testing.T
	name         string
	expectedDeps map[string]any
	ok           bool
	res          any
	err          error
	delay        time.Duration
}

func newMockDataSourceProvider(delay time.Duration) *mockDataSourceProvider {
	return &mockDataSourceProvider{
		name2Getter: make(map[string]DataSourceGetter),
		delay:       delay,
	}
}

func (m *mockDataSourceProvider) addGet(getter *mockGetter) {
	getter.delay = m.delay
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
	if g.delay != 0 {
		timer := time.NewTimer(g.delay)

		defer timer.Stop()
		select {
		case <-dctx.Ctx.Done():
			return nil, dctx.Ctx.Err()
		case <-timer.C:
		}
	}

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

func newTestCacheServiceConfig(timeout time.Duration) *config.CacheServiceConfig {
	return &config.CacheServiceConfig{
		DataSources: []*config.DataSourceConfig{
			newSimpleDataSourceConfig("A", nil),
			newSimpleDataSourceConfig("B", []string{"A"}),
			newSimpleDataSourceConfig("C", []string{"A"}),
			newSimpleDataSourceConfig("D", []string{"B"}),
			newSimpleDataSourceConfig("E", []string{"B", "C"}),
		},
		GlobalTimeout:  common.Duration{Duration: timeout},
		MaxParallelism: -1,
	}
}

//     A
//    / \
//   B   C
//  / \ /
// D   E

func TestFetchOk(t *testing.T) {
	cfg := newTestCacheServiceConfig(time.Second * 10)
	provider := newMockDataSourceProvider(0)
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

func TestFetchFail(t *testing.T) {
	berr := errors.New("Berr")

	cfg := newTestCacheServiceConfig(time.Second)
	provider := newMockDataSourceProvider(0)
	provider.addOkGetter(t, "A", "Ares", map[string]any{})
	provider.addFailGetter(t, "B", berr, map[string]any{"A": "Ares"})
	provider.addOkGetter(t, "C", "Cres", map[string]any{"A": "Ares"})
	provider.addOkGetter(t, "D", "Dres", map[string]any{"B": "Bres"})
	provider.addOkGetter(t, "E", "Eres", map[string]any{"B": "Bres", "C": "Cres"})

	cacheService, err := NewCacheService(cfg, provider)
	if err != nil {
		t.Fatalf("expected no err, got: %v", err)
	}

	_, err = cacheService.GetOrderInfo(context.Background(), "kek", "lol")

	if err != berr {
		t.Errorf("unexpected error, expected: %v, got: %v", berr, err)
	}
}

func TestFetchTimeout(t *testing.T) {
	delay := time.Millisecond * 200
	cfg := newTestCacheServiceConfig(delay / 2)
	provider := newMockDataSourceProvider(delay)
	provider.addOkGetter(t, "A", "Ares", map[string]any{})
	provider.addOkGetter(t, "B", "Bres", map[string]any{"A": "Ares"})
	provider.addOkGetter(t, "C", "Cres", map[string]any{"A": "Ares"})
	provider.addOkGetter(t, "D", "Dres", map[string]any{"B": "Bres"})
	provider.addOkGetter(t, "E", "Eres", map[string]any{"B": "Bres", "C": "Cres"})

	cacheService, err := NewCacheService(cfg, provider)
	if err != nil {
		t.Fatalf("expected no err, got: %v", err)
	}

	_, err = cacheService.GetOrderInfo(context.Background(), "kek", "lol")
	if err == nil {
		t.Fatalf("expected err")
	}
}

func TestFetchContextTimeout(t *testing.T) {
	delay := time.Millisecond * 200
	cfg := newTestCacheServiceConfig(delay)
	provider := newMockDataSourceProvider(delay * 2)
	provider.addOkGetter(t, "A", "Ares", map[string]any{})
	provider.addOkGetter(t, "B", "Bres", map[string]any{"A": "Ares"})
	provider.addOkGetter(t, "C", "Cres", map[string]any{"A": "Ares"})
	provider.addOkGetter(t, "D", "Dres", map[string]any{"B": "Bres"})
	provider.addOkGetter(t, "E", "Eres", map[string]any{"B": "Bres", "C": "Cres"})

	cacheService, err := NewCacheService(cfg, provider)
	if err != nil {
		t.Fatalf("expected no err, got: %v", err)
	}

	ctx, close := context.WithTimeout(context.Background(), delay/2)
	defer close()

	_, err = cacheService.GetOrderInfo(ctx, "kek", "lol")
	if err == nil {
		t.Fatalf("expected err")
	}
}
