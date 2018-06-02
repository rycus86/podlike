package engine

import (
	"context"
	"github.com/docker/docker/api/types"
	"io"
)

func (e *Engine) CopyToContainer(containerID string, destPath string, content io.Reader) error {
	// TODO proper context
	return e.api.CopyToContainer(context.TODO(), containerID, destPath, content, types.CopyToContainerOptions{})
}
