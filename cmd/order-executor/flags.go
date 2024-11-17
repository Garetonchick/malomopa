package main

import "flag"

type OrderExecutorFlags struct {
	configPath *string
}

func parseFlags() *OrderExecutorFlags {
	configPath := flag.String("config", "", "Path to order executor config")

	flag.Parse()

	return &OrderExecutorFlags{
		configPath: configPath,
	}
}
