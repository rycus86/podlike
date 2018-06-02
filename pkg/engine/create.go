package engine

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"time"
)

func (e *Engine) CreateContainer(config *container.Config, hostConfig *container.HostConfig, name string) (container.ContainerCreateCreatedBody, error) {
	ctxCreate, cancelCreate := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelCreate()

	return e.api.ContainerCreate(ctxCreate,
		config,
		hostConfig,
		&network.NetworkingConfig{},
		name)
}
