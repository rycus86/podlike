package engine

import (
	"context"
	"github.com/docker/docker/api/types"
	"io"
)

func (e *Engine) StreamLogs(containerID string) (io.ReadCloser, error) {
	return e.api.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
}
