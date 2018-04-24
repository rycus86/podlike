package config

import "flag"

type Configuration struct {
	SharePids    bool
	ShareVolumes bool
	StreamLogs   bool
	AlwaysPull   bool
}

func Parse() *Configuration {
	var (
		pids, volumes, logs, pull bool
	)

	flag.BoolVar(&pids, "pids", true, "Enable (default) or disable PID sharing")
	flag.BoolVar(&volumes, "volumes", true, "Enable (default) or disable volume sharing")
	flag.BoolVar(&logs, "logs", false, "Stream logs from the components")
	flag.BoolVar(&logs, "pull", false, "Always pull the images for the components when starting")
	flag.Parse()

	return &Configuration{
		SharePids:    pids,
		ShareVolumes: volumes,
		StreamLogs:   logs,
		AlwaysPull:   pull,
	}
}
