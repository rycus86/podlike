package engine

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/rycus86/podlike/pkg/config"
)

type Engine struct {
	api *client.Client
	auth *config.RegistryAuth

	cancelEvents context.CancelFunc
}
