package engine

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
)

func (e *Engine) WatchHealthcheckEvents() (<-chan events.Message, <-chan error) {
	ctx, cancel := context.WithCancel(context.Background())
	e.cancelEvents = cancel

	return e.api.Events(ctx, types.EventsOptions{
		Filters: filters.NewArgs(filters.Arg("event", "health_status")),
	})
}
