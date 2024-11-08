package main

import (
	"log"
	cacheservice "malomopa/internal/cache-service"
	"malomopa/internal/config"
	calc "malomopa/internal/cost-calculator"
	"malomopa/internal/db"
	assigner "malomopa/internal/order-assigner"

	"golang.org/x/sync/errgroup"
)

func main() {
	inputFlags := parseFlags()

	configPathP := inputFlags.configPath
	if configPathP == nil || *configPathP == "" {
		log.Fatal("No config file provided")
	}
	configPath := *configPathP

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	cacheServiceProvider, err := cacheservice.MakeCacheService(cfg.CacheService)
	if err != nil {
		log.Fatal(err)
	}

	dbProvider, err := db.MakeDBProvider(cfg.Scylla)
	if err != nil {
		log.Fatal(err)
	}

	costCalculator, err := calc.MakeSimpleCostCalculator()
	if err != nil {
		log.Fatal(err)
	}

	server, err := assigner.NewServer(
		cfg.HTTPServer,
		cacheServiceProvider,
		dbProvider,
		costCalculator,
	)
	if err != nil {
		log.Fatal(err)
	}

	var wg errgroup.Group
	wg.Go(func() error {
		return server.Run()
	})

	err = wg.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
