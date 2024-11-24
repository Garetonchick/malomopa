package cacheservice

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"malomopa/internal/common"
	"malomopa/internal/config"

	"github.com/karlseguin/ccache/v3"
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

	// TODO: proper data sources config + initialization
	getGeneralOrderInfoF = registerFetcher(registerFetcherCfg{
		get:      getGeneralOrderInfo,
		name:     common.GeneralOrderInfoKey,
		endpoint: cfg.GetGeneralOrderInfoEndpoint,
		deps:     nil,
		cacheCfg: nil,
	})

	getZoneInfoF = registerFetcher(registerFetcherCfg{
		get:      getZoneInfo,
		name:     common.ZoneInfoKey,
		endpoint: cfg.GetZoneInfoEndpoint,
		deps:     []*fetcher{getGeneralOrderInfoF},
		cacheCfg: &fetcherCacheConfig{maxSize: 1000, ttl: time.Minute * 1},
	})

	getExecutorProfileF = registerFetcher(registerFetcherCfg{
		get:      getExecutorProfile,
		name:     common.ExecutorProfileKey,
		endpoint: cfg.GetExecutorProfileEndpoint,
		deps:     nil,
		cacheCfg: nil,
	})

	getConfigsF = registerFetcher(registerFetcherCfg{
		get:      getConfigs,
		name:     common.ConfigsKey,
		endpoint: cfg.GetConfigsEndpoint,
		deps:     nil,
		cacheCfg: &fetcherCacheConfig{maxSize: 1, ttl: time.Minute * 1},
	})

	getTollRoadsInfoF = registerFetcher(registerFetcherCfg{
		get:      getTollRoadsInfo,
		name:     common.TollRoadsInfoKey,
		endpoint: cfg.GetTollRoadsInfoEndpoint,
		deps:     []*fetcher{getZoneInfoF},
		cacheCfg: nil,
	})

	return &cacheService, nil
}

// /////////////////////////////////////////////////////////////////////////////

type fetcherCacheConfig struct {
	maxSize int64
	ttl     time.Duration
}

type fetcherCache struct {
	cfg   *fetcherCacheConfig
	cache *ccache.Cache[any] // Is any really ok here?
}

func (fc *fetcherCache) Get(key string) any {
	res := fc.cache.Get(key)
	if res != nil && !res.Expired() {
		val := res.Value()
		return &val
	}
	return nil
}

func (fc *fetcherCache) Set(key string, value any) {
	fc.cache.Set(key, value, fc.cfg.ttl)
}

func MakeFetcherCache(cfg *fetcherCacheConfig) *fetcherCache {
	if cfg == nil {
		return nil
	}
	return &fetcherCache{
		cfg:   cfg,
		cache: ccache.New(ccache.Configure[any]().MaxSize(cfg.maxSize)),
	}
}

///////////////////////////////////////////////////////////////////////////////

type fetcherID uint64

type call struct {
	Ctx        context.Context
	OrderID    string
	ExecutorID string
}

type fetcherFunc func(*call, *fetcherCache, string, map[fetcherID]any) (any, error)

type fetcher struct {
	Get      fetcherFunc
	GetCache *fetcherCache
	ID       fetcherID
	Name     string
	Endpoint string
	Deps     []*fetcher
}

type registerFetcherCfg struct {
	get      fetcherFunc
	name     string
	endpoint string
	timeout  time.Duration // 0 is inf
	deps     []*fetcher
	cacheCfg *fetcherCacheConfig
}

type job struct {
	Fetcher  *fetcher
	Parents  []*job
	DepsLeft atomic.Int32
	Result   any
	Error    error
}

var fetchers = []fetcher{}

func registerFetcher(cfg registerFetcherCfg) *fetcher {
	// TODO: burn with
	getWithTimeout := func(c *call, cache *fetcherCache, endpoint string, deps map[fetcherID]any) (any, error) {
		if cfg.timeout == 0 {
			return cfg.get(c, cache, endpoint, deps)
		}

		newC := c
		ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
		newC.Ctx = ctx
		defer cancel()
		resChan := make(chan any)
		errChan := make(chan error)
		go func() {
			res, err := cfg.get(newC, cache, endpoint, deps)
			if err != nil {
				errChan <- err
			} else {
				resChan <- res
			}
		}()
		var res any
		var err error
		select {
		case <-c.Ctx.Done():
			errMsg := fmt.Sprintf("Timeout expired for fetcher '%s'", cfg.name)
			fmt.Println(errMsg)
			return nil, errors.New(errMsg)
		case res = <-resChan:
			return res, nil
		case err = <-errChan:
			return nil, err
		}
	}

	f := fetcher{
		Get:      getWithTimeout,
		GetCache: MakeFetcherCache(cfg.cacheCfg),
		ID:       fetcherID(len(fetchers)),
		Name:     cfg.name,
		Endpoint: cfg.endpoint,
		Deps:     cfg.deps,
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

		res, err := jb.Fetcher.Get(&c, jb.Fetcher.GetCache, jb.Fetcher.Endpoint, deps)
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
