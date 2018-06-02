package engine

import (
	"context"
	"github.com/docker/docker/api/types"
	"time"
)

func (e *Engine) StartContainer(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return e.api.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}
