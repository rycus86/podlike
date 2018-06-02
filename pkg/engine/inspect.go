package engine

import (
	"context"
	"github.com/docker/docker/api/types"
	"time"
)

func (e *Engine) InspectContainer(containerID string) (*types.ContainerJSON, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	container, err := e.api.ContainerInspect(ctx, containerID)
	return &container, err
}

func (e *Engine) InspectVolume(name string) (types.Volume, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return e.api.VolumeInspect(ctx, name)
}
