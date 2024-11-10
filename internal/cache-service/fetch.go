package cacheservice

import (
	"context"
	"errors"
	"log"
	"malomopa/internal/common"
	"malomopa/internal/config"
	"sync"
	"sync/atomic"
)

type CacheService struct {
	cfg *config.CacheServiceConfig
}

var getGeneralOrderInfoF *fetcher
var getZoneInfoF *fetcher
var getExecutorProfileF *fetcher
var getConfigsF *fetcher
var getTollRoadsInfoF *fetcher

var ErrCacheServiceMisconfigured error = errors.New("cache service misconfigured")

func MakeCacheService(cfg *config.CacheServiceConfig) (common.CacheServiceProvider, error) {
	if cfg == nil {
		return nil, ErrCacheServiceMisconfigured
	}

	cacheService := CacheService{
		cfg: cfg,
	}

	_ = getExecutorProfileF
	_ = getConfigsF
	_ = getTollRoadsInfoF

	getGeneralOrderInfoF = registerFetcher(
		getGeneralOrderInfo, common.GeneralOrderInfoKey, cfg.GetGeneralOrderInfoEndpoint, nil,
	)
	getZoneInfoF = registerFetcher(
		getZoneInfo, common.ZoneInfoKey, cfg.GetZoneInfoEndpoint, []*fetcher{getGeneralOrderInfoF},
	)
	getExecutorProfileF = registerFetcher(
		getExecutorProfile, common.ExecutorProfileKey, cfg.GetExecutorProfileEndpoint, nil,
	)
	getConfigsF = registerFetcher(
		getConfigs, common.ConfigsKey, cfg.GetConfigsEndpoint, nil,
	)
	getTollRoadsInfoF = registerFetcher(
		getTollRoadsInfo, common.TollRoadsInfoKey, cfg.GetTollRoadsInfoEndpoint, []*fetcher{getZoneInfoF},
	)

	return &cacheService, nil
}

type fetcherID uint64

type call struct {
	Ctx        context.Context
	OrderID    string
	ExecutorID string
}

type fetcherFunc func(*call, string, map[fetcherID]any) (any, error)

type fetcher struct {
	Get      fetcherFunc
	ID       fetcherID
	Name     string
	Endpoint string
	Deps     []*fetcher
}

type job struct {
	Fetcher  *fetcher
	Parents  []*job
	DepsLeft atomic.Int32
	Result   any
	Error    error
}

var fetchers = []fetcher{}

func registerFetcher(get fetcherFunc, name, endpoint string, deps []*fetcher) *fetcher {
	f := fetcher{
		Get:      get,
		ID:       fetcherID(len(fetchers)),
		Name:     name,
		Endpoint: endpoint,
		Deps:     deps,
	}
	fetchers = append(fetchers, f)
	return &fetchers[f.ID]
}

func buildJobsGraph() []job {
	var jobs = make([]job, len(fetchers))

	var dfs func(f *fetcher)
	dfs = func(f *fetcher) {
		jobs[f.ID].Fetcher = f
		jobs[f.ID].DepsLeft.Store(int32(len(f.Deps)))

		for _, dep := range f.Deps {
			if jobs[dep.ID].Fetcher == nil {
				dfs(dep)
			}
			jobs[dep.ID].Parents = append(jobs[dep.ID].Parents, &jobs[f.ID])
		}
	}

	for i := range fetchers {
		if jobs[i].Fetcher == nil {
			dfs(&fetchers[i])
		}
	}

	return jobs
}

func (cs *CacheService) GetOrderInfo(ctx context.Context, orderID string, executorID string) (common.OrderInfo, error) {
	c := call{
		Ctx:        ctx,
		OrderID:    orderID,
		ExecutorID: executorID,
	}

	wg := sync.WaitGroup{}
	jobs := buildJobsGraph()

	var worker func(jb *job)
	worker = func(jb *job) {
		defer wg.Done()

		deps := make(map[fetcherID]any)

		for _, dep := range jb.Fetcher.Deps {
			deps[dep.ID] = jobs[dep.ID].Result
		}

		res, err := jb.Fetcher.Get(&c, jb.Fetcher.Endpoint, deps)
		jb.Result = res
		jb.Error = err

		if err != nil {
			return
		}

		var execs []*job

		for _, p := range jb.Parents {
			if p.DepsLeft.Add(-1) == 0 {
				execs = append(execs, p)
			}
		}

		for i := 0; i+1 < len(execs); i += 1 {
			wg.Add(1)
			go worker(execs[i])
		}

		if len(execs) > 0 {
			wg.Add(1)
			worker(execs[len(execs)-1])
		}
	}

	for i := range jobs {
		if len(fetchers[i].Deps) == 0 {
			wg.Add(1)
			go worker(&jobs[i])
		}
	}

	wg.Wait()

	name2data := make(map[string]any)
	var err error

	for i := range jobs {
		if jobs[i].Result != nil && jobs[i].Error == nil {
			name2data[jobs[i].Fetcher.Name] = jobs[i].Result
		} else if jobs[i].Error != nil {
			log.Printf(
				"fetching from source %q failed: %s",
				jobs[i].Fetcher.Name,
				jobs[i].Error,
			)
			err = errors.New("fetching sources error")
		} else {
			log.Printf(
				"skipping fetching of %q data source because some dependencies failed",
				jobs[i].Fetcher.Name,
			)
			err = errors.New("fetching sources error")
		}
	}

	return name2data, err
}
