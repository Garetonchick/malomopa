package main

import (
	"errors"
	cacheservice "malomopa/internal/cache-service"
	common "malomopa/internal/common"
	"malomopa/internal/config"
	calc "malomopa/internal/cost-calculator"
	"malomopa/internal/db"
	assigner "malomopa/internal/order-assigner"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	inputFlags := parseFlags()

	configPathP := inputFlags.configPath
	if configPathP == nil || *configPathP == "" {
		common.TerminateWithErr(errors.New("no config file provided"))
	}
	configPath := *configPathP

	cfg, err := config.LoadAssignerConfig(configPath)
	if err != nil {
		common.TerminateWithErr(err)
	}

	logger, err := config.MakeLogger(cfg.Logger)
	if err != nil {
		common.TerminateWithErr(err)
	}

	cacheServiceProvider, err := cacheservice.NewCacheService(
		cfg.CacheService,
		cacheservice.NewDataSourcesProvider(),
	)

	if err != nil {
		logger.Fatal("failed to create cache service provider",
			zap.String("err", err.Error()),
		)
	}
	logger.Info("Cache Service configured successfuly")

	dbProvider, err := db.MakeDBProvider(cfg.Scylla)
	if err != nil {
		logger.Fatal("failed to create DB provider",
			zap.String("err", err.Error()),
		)
	}
	logger.Info("DB configured successfuly")

	costCalculator, err := calc.MakeSimpleCostCalculator()
	if err != nil {
		logger.Fatal("failed to create cost calculator",
			zap.String("err", err.Error()),
		)
	}
	logger.Info("Cost Calculator configured successfuly")

	server, err := assigner.NewServer(
		cfg.HTTPServer,
		cacheServiceProvider,
		dbProvider,
		costCalculator,
		logger,
	)
	if err != nil {
		logger.Fatal("failed to create http server",
			zap.String("err", err.Error()),
		)
	}
	logger.Info("HTTP Server configured successfuly")

	var wg errgroup.Group
	wg.Go(func() error {
		return server.Run(logger)
	})

	err = wg.Wait()
	if err != nil {
		logger.Fatal("http server exited",
			zap.String("err", err.Error()),
		)
	}
}
