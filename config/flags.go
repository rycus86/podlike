package config

import "flag"

type Configuration struct {
	SharePids    bool
	ShareVolumes bool
}

func Parse() *Configuration {
	var (
		pids, volumes bool
	)

	flag.BoolVar(&pids, "pids", true, "Enable (default) or disable PID sharing")
	flag.BoolVar(&volumes, "volumes", true, "Enable (default) or disable volume sharing")
	flag.Parse()

	return &Configuration{
		SharePids:    pids,
		ShareVolumes: volumes,
	}
}
