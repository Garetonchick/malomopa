package cacheservice

import (
	"context"
	"errors"
	"malomopa/internal/config"
	"reflect"
	"testing"
)

type mockGet struct {
	Name         string
	Endpoint     string
	Ok           bool
	Res          any
	Err          error
	ExpectedDeps map[fetcherID]any
	T            *testing.T
	GetCache     *fetcherCache
}

func (m *mockGet) Get(_ *call, _ *fetcherCache, _ string, deps map[fetcherID]any) (any, error) {
	if m.ExpectedDeps != nil && !reflect.DeepEqual(deps, m.ExpectedDeps) {
		m.T.Errorf("Expected %v but got %v deps for %s", m.ExpectedDeps, deps, m.Name)
	}
	if m.Ok {
		return m.Res, nil
	} else {
		return nil, m.Err
	}
}

func (m *mockGet) Register(deps []*fetcher) *fetcher {
	return registerFetcher(m.Get, m.Name, m.Endpoint, deps, nil)
}

func newOkGet(t *testing.T, name string, res any, expected map[fetcherID]any) *mockGet {
	return &mockGet{
		T:            t,
		Name:         name,
		Ok:           true,
		Res:          res,
		ExpectedDeps: expected,
	}
}

func newFailGet(t *testing.T, name string, err error, expected map[fetcherID]any) *mockGet {
	return &mockGet{
		T:            t,
		Name:         name,
		Ok:           false,
		Err:          err,
		ExpectedDeps: expected,
	}
}

func resetFetchers() (restore func()) {
	oldFetchers := fetchers
	fetchers = make([]fetcher, 0)
	restore = func() {
		fetchers = oldFetchers
	}
	return restore
}

//	   A
//	  / \
//	 B   C
//	/ \ /
// D   E

func TestFetchingOk(t *testing.T) {
	cfg := &config.CacheServiceConfig{}
	cacheService, _ := MakeCacheService(cfg)
	restore := resetFetchers()
	defer restore()
	getA := newOkGet(t, "A", "Ares", map[fetcherID]any{})
	aF := getA.Register(nil)
	getB := newOkGet(t, "B", "Bres", map[fetcherID]any{aF.ID: "Ares"})
	bF := getB.Register([]*fetcher{aF})
	getC := newOkGet(t, "C", "Cres", map[fetcherID]any{aF.ID: "Ares"})
	cF := getC.Register([]*fetcher{aF})
	getD := newOkGet(t, "D", "Dres", map[fetcherID]any{bF.ID: "Bres"})
	_ = getD.Register([]*fetcher{bF})
	getE := newOkGet(t, "E", "Eres", map[fetcherID]any{bF.ID: "Bres", cF.ID: "Cres"})
	_ = getE.Register([]*fetcher{bF, cF})

	fetched, err := cacheService.GetOrderInfo(context.Background(), "kek", "lol")
	if err != nil {
		t.Errorf("expected: nil, got: %v", err)
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

//	   A
//	  / \
//	 B*  C
//	/ \ /
// D   E

func TestFetchingFailures(t *testing.T) {
	cfg := &config.CacheServiceConfig{}
	cacheService, _ := MakeCacheService(cfg)
	restore := resetFetchers()
	defer restore()
	getA := newOkGet(t, "A", "Ares", map[fetcherID]any{})
	aF := getA.Register(nil)
	getB := newFailGet(t, "B", errors.New("Berr"), map[fetcherID]any{aF.ID: "Ares"})
	bF := getB.Register([]*fetcher{aF})
	getC := newOkGet(t, "C", "Cres", map[fetcherID]any{aF.ID: "Ares"})
	cF := getC.Register([]*fetcher{aF})
	getD := newOkGet(t, "D", "Dres", map[fetcherID]any{bF.ID: "Bres"})
	_ = getD.Register([]*fetcher{bF})
	getE := newOkGet(t, "E", "Eres", map[fetcherID]any{bF.ID: "Bres", cF.ID: "Cres"})
	_ = getE.Register([]*fetcher{bF, cF})

	fetched, err := cacheService.GetOrderInfo(context.Background(), "kek", "lol")
	if err == nil {
		t.Errorf("expected: not nil, got: nil")
	}

	expected := map[string]string{
		"A": "Ares",
		"C": "Cres",
	}

	if !compareJSONs(fetched, expected) {
		t.Errorf("expected: %v, got: %v", expected, fetched)
	}
}
