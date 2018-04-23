package config

import "flag"

type Configuration struct {
	SharePids    bool
	ShareVolumes bool
	StreamLogs   bool
}

func Parse() *Configuration {
	var (
		pids, volumes, logs bool
	)

	flag.BoolVar(&pids, "pids", true, "Enable (default) or disable PID sharing")
	flag.BoolVar(&volumes, "volumes", true, "Enable (default) or disable volume sharing")
	flag.BoolVar(&logs, "logs", false, "Stream logs from the components")
	flag.Parse()

	return &Configuration{
		SharePids:    pids,
		ShareVolumes: volumes,
		StreamLogs:   logs,
	}
}
