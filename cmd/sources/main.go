package main

import (
	"flag"
	"log"
	common "malomopa/internal/common"
	sources "malomopa/internal/sources"
)

func main() {
	configPath := flag.String("config", "", "Path to order assigner config")
	flag.Parse()
	if configPath == nil {
		log.Fatal("no config file provided")
	}

	cfg, err := common.ReadJSONFromFile[sources.Config](*configPath)
	if err != nil {
		log.Fatal("config has wrong format or doesn't exist")
	}

	sources.FakeInfo, err = sources.NewFakeInfo(cfg.DataPaths)
	if err != nil {
		log.Fatal("fake info initiation failed :", err.Error())
	}

	s, err := sources.NewServer(cfg.HttpServer)
	if err != nil {
		log.Fatal("something gone wrong: ", err.Error())
	}

	if err := s.Run(); err != nil {
		log.Fatal("`server.Run()` finished with error: ", err.Error())
	}
}
