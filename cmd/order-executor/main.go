package main

import (
	"errors"
	"fmt"
	"malomopa/internal/config"
	"malomopa/internal/db"
	acquirer "malomopa/internal/order-executor"
	"os"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func terminateWithErr(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func main() {
	inputFlags := parseFlags()

	configPathP := inputFlags.configPath
	if configPathP == nil || *configPathP == "" {
		terminateWithErr(errors.New("no config file provided"))
	}
	configPath := *configPathP

	cfg, err := config.LoadAcquirerConfig(configPath)
	if err != nil {
		terminateWithErr(err)
	}

	logger, err := config.MakeLogger(cfg.Logger)
	if err != nil {
		terminateWithErr(err)
	}

	dbProvider, err := db.MakeDBProvider(cfg.Scylla)
	if err != nil {
		logger.Fatal("failed to create DB provider",
			zap.String("err", err.Error()),
		)
	}

	server, err := acquirer.NewServer(
		cfg,
		dbProvider,
		logger,
	)
	if err != nil {
		logger.Fatal("failed to create http server",
			zap.String("err", err.Error()),
		)
	}

	var wg errgroup.Group
	wg.Go(func() error {
		return server.Run()
	})

	err = wg.Wait()
	if err != nil {
		logger.Fatal("http server exited",
			zap.String("err", err.Error()),
		)
	}
}
