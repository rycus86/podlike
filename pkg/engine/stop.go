package engine

import (
	"context"
	"github.com/docker/docker/api/types"
	"time"
)

func (e *Engine) StopContainer(containerID string, timeout *time.Duration) error {
	var contextTimeout time.Duration

	if timeout != nil {
		contextTimeout = *timeout + 1*time.Second
	} else {
		contextTimeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	return e.api.ContainerStop(ctx, containerID, timeout)
}

func (e *Engine) RemoveContainer(containerID string) error {
	return e.api.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
		Force: true,
	})
}
