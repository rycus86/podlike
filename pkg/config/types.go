package config

import "github.com/docker/docker/api/types"

type Configuration struct {
	SharePids    bool
	ShareIpc     bool
	ShareVolumes bool
	StreamLogs   bool
	AlwaysPull   bool
}

type RegistryAuth struct {
	Auths map[string]types.AuthConfig `json:"auths"`
}