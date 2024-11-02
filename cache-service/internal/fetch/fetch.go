package fetch

import (
	"context"
	"errors"
)

type fetcherID uint64

type call struct {
	Ctx        context.Context
	OrderID    string
	ExecutorID string
}

type fetcherFunc func(*call, map[fetcherID]any) (any, error)

type fetcher struct {
	Get  fetcherFunc
	ID   fetcherID
	Name string
	Deps []*fetcher
}

var fetchers = []fetcher{}

func registerFetcher(get fetcherFunc, name string, deps []*fetcher) *fetcher {
	f := fetcher{
		Get:  get,
		ID:   fetcherID(len(fetchers)),
		Name: name,
		Deps: deps,
	}
	fetchers = append(fetchers, f)
	return &fetchers[f.ID]
}

func TryAll(ctx context.Context, orderID string, executorID string) map[string]any {
	c := call{
		Ctx:        ctx,
		OrderID:    orderID,
		ExecutorID: executorID,
	}

	results := make(map[fetcherID]any)
	var dfs func(f *fetcher) (any, error)
	dfs = func(f *fetcher) (any, error) {
		results[f.ID] = nil

		deps := make(map[fetcherID]any)

		for _, dep := range f.Deps {
			if res, ok := results[dep.ID]; ok {
				deps[dep.ID] = res
			} else if res == nil {
				return nil, errors.New("dependency data source is not available")
			}
			res, err := dfs(dep)
			if err != nil {
				return nil, err
			}
			deps[dep.ID] = res
		}

		res, err := f.Get(&c, deps)
		if err != nil {
			return nil, err
		}
		results[f.ID] = res

		return res, nil
	}

	name2data := make(map[string]any)

	for _, f := range fetchers {
		res, err := dfs(&f)
		if err == nil && res != nil {
			name2data[f.Name] = res
		}
	}

	return name2data
}
