package engine

import (
	"context"
	"github.com/docker/docker/api/types/container"
)

func (e *Engine) WaitContainer(containerID string) (<-chan container.ContainerWaitOKBody, <-chan error) {
	return e.api.ContainerWait(context.Background(), containerID, container.WaitConditionNotRunning)
}
