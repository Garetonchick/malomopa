package main

import (
	"errors"
	common "malomopa/internal/common"
	"malomopa/internal/config"
	"malomopa/internal/db"
	executor "malomopa/internal/order-executor"

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

	cfg, err := config.LoadExecutorConfig(configPath)
	if err != nil {
		common.TerminateWithErr(err)
	}

	logger, err := config.MakeLogger(cfg.Logger)
	if err != nil {
		common.TerminateWithErr(err)
	}

	dbProvider, err := db.MakeDBProvider(cfg.Scylla)
	if err != nil {
		logger.Fatal("failed to create DB provider",
			zap.String("err", err.Error()),
		)
	}
	logger.Info("DB configured successfuly")

	server, err := executor.NewServer(
		cfg,
		dbProvider,
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
