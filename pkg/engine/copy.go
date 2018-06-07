package engine

import (
	"context"
	"github.com/docker/docker/api/types"
	"io"
)

func (e *Engine) CopyToContainer(containerID string, destPath string, content io.Reader) error {
	// TODO is context.Background() appropriate here?
	return e.api.CopyToContainer(context.Background(), containerID, destPath, content, types.CopyToContainerOptions{})
}
