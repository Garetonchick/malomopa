package main

import "flag"

type OrderAssignerFlags struct {
	configPath *string
}

func parseFlags() *OrderAssignerFlags {
	configPath := flag.String("config", "", "Path to order assigner config")

	return &OrderAssignerFlags{
		configPath: configPath,
	}
}
