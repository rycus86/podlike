package config

type Configuration struct {
	SharePids    bool
	ShareIpc     bool
	ShareVolumes bool
	StreamLogs   bool
	AlwaysPull   bool
}
