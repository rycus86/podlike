package engine

import (
	"context"
	"github.com/docker/docker/client"
)

type Engine struct {
	api *client.Client

	cancelEvents context.CancelFunc
}
