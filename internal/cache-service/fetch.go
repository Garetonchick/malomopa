package cacheservice

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"malomopa/internal/common"
	"malomopa/internal/config"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const cacheServiceName = "cache_service"

type cacheService struct {
	cfg   *config.CacheServiceConfig
	graph []*sourceNode
}

type sourceNode struct {
	index int
	name  string

	getter DataSourceGetter

	cache    Cache
	timeout  *time.Duration
	endpoint string

	deps []*sourceNode
}

var ErrCacheServiceMisconfigured error = errors.New("cache service misconfigured")

func NewCacheService(
	cfg *config.CacheServiceConfig, provider DataSourcesProvider,
) (common.CacheServiceProvider, error) {
	if cfg == nil {
		return nil, ErrCacheServiceMisconfigured
	}

	sources, err := scanSources(provider, cfg.DataSources)
	if err != nil {
		return nil, err
	}

	err = linkSources(sources, cfg.DataSources)
	if err != nil {
		return nil, err
	}

	return &cacheService{
		cfg:   cfg,
		graph: sources,
	}, nil
}

func scanSources(provider DataSourcesProvider, cfgs []*config.DataSourceConfig) ([]*sourceNode, error) {
	var sources []*sourceNode

	for i, source := range cfgs {
		getter, err := provider.GetGet(source.Name)
		if err != nil {
			return nil, err
		}
		cache, err := NewCache(source.Cache)
		if err != nil {
			return nil, err
		}

		sources = append(
			sources,
			&sourceNode{
				index: i,
				name:  source.Name,

				getter:   getter,
				cache:    cache,
				timeout:  source.Timeout,
				endpoint: source.Endpoint,

				deps: nil,
			},
		)
	}
	return sources, nil
}

func linkSources(sources []*sourceNode, cfgs []*config.DataSourceConfig) error {
	name2Node := make(map[string]*sourceNode)

	for _, s := range sources {
		name2Node[s.name] = s
	}

	for _, cfg := range cfgs {
		node, ok := name2Node[cfg.Name]
		if !ok {
			panic("bug in scanSources")
		}
		for _, dep := range cfg.Deps {
			depNode, ok := name2Node[dep]
			if !ok {
				return fmt.Errorf("dependency node %q not found in sources list", dep)
			}
			node.deps = append(node.deps, depNode)
		}
	}
	return nil
}

type jobNode struct {
	source   *sourceNode
	parents  []*jobNode
	depsLeft atomic.Int32
	result   any
}

func buildJobsGraph(sources []*sourceNode) []*jobNode {
	var jobs = make([]*jobNode, len(sources))
	for i := range jobs {
		jobs[i] = &jobNode{}
	}

	var dfs func(s *sourceNode)
	dfs = func(s *sourceNode) {
		job := jobs[s.index]
		job.source = s
		job.depsLeft.Store(int32(len(s.deps)))

		for _, dep := range s.deps {
			if jobs[dep.index].source == nil {
				dfs(dep)
			}
			jobs[dep.index].parents = append(jobs[dep.index].parents, job)
		}
	}

	for i := range sources {
		if jobs[i].source == nil {
			dfs(sources[i])
		}
	}

	return jobs
}

func (cs *cacheService) GetOrderInfo(
	ctx context.Context, orderID string, executorID string,
) (common.OrderInfo, error) {
	logger := common.GetRequestLogger(ctx, cacheServiceName, "get_order_info")

	ctx, cancel := context.WithTimeout(ctx, cs.cfg.GlobalTimeout)
	defer cancel()

	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(cs.cfg.MaxParallelism) // Semaphore

	logger.Info("start building jobs graph",
		zap.String("order_id", orderID),
		zap.String("executor_id", executorID),
	)

	jobs := buildJobsGraph(cs.graph)

	jobsChan := make(chan *jobNode, len(jobs))

	for i, job := range jobs {
		// add to initial runqueue if it is a leaf node
		if len(cs.graph[i].deps) == 0 {
			jobsChan <- job
		}
	}

	logger.Info("built jobs graph, start fetching",
		zap.Int("n_jobs", len(jobs)),
	)
	fetchStartTime := time.Now()

	req := DataSourcesRequest{
		OrderID:    orderID,
		ExecutorID: executorID,
		Logger:     logger,
	}

	var loopErr error

	for range len(jobs) {
		var job *jobNode

		select {
		case job = <-jobsChan:
		case <-ctx.Done():
		}

		if ctx.Err() != nil {
			loopErr = ctx.Err()
			break
		}

		wg.Go(func() error {
			return processJob(
				ctx,
				&req,
				jobs,
				job,
				jobsChan,
			)
		})
	}

	err := wg.Wait()
	if err == nil {
		err = loopErr
	}
	if err != nil {
		logger.Info("failed to fetch some data sources")
		return nil, err
	}

	logger.Info("fetched all data sources",
		zap.Duration("total_fetch_time", time.Since(fetchStartTime)),
	)

	return common.OrderInfo(collectJobResults(jobs)), nil
}

func processJob(
	ctx context.Context, req *DataSourcesRequest, jobs []*jobNode, curJob *jobNode, ch chan *jobNode,
) error {
	deps := make(map[string]any)
	source := curJob.source

	for _, dep := range source.deps {
		deps[dep.name] = jobs[dep.index].result
	}

	var cancel context.CancelFunc
	if source.timeout != nil {
		ctx, cancel = context.WithTimeout(ctx, *source.timeout)
		defer cancel()
	}

	res, err := source.getter.Get(req, &DataSourceContext{
		Ctx:      ctx,
		Cache:    source.cache,
		Endpoint: source.endpoint,
		Deps:     deps,
	})
	if err != nil {
		return err
	}

	curJob.result = res

	for _, p := range curJob.parents {
		if p.depsLeft.Add(-1) == 0 {
			ch <- p
		}
	}

	return nil
}

func collectJobResults(jobs []*jobNode) map[string]any {
	name2data := make(map[string]any)

	for i := range jobs {
		name2data[jobs[i].source.name] = jobs[i].result
	}

	return name2data
}
