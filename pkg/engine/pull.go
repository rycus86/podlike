package engine

import (
	"context"
	"github.com/docker/docker/api/types"
	"io"
)

func (e *Engine) PullImage(reference string) (io.ReadCloser, error) {
	// TODO is context.Background() appropriate here?
	return e.api.ImagePull(context.Background(), reference, types.ImagePullOptions{})
}
