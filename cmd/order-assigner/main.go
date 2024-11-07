package main

import (
	"log"
	"malomopa/internal/config"
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

	server, err := assigner.NewServer(cfg)
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
