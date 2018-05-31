package engine

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/rycus86/podlike/healthcheck"
	"strings"
)

func (c *Client) WatchHealthcheckEvents() {
	for {
		if c.closed {
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		c.cancelEvents = cancel

		chMessage, chErr := c.api.Events(ctx, types.EventsOptions{
			Filters: filters.NewArgs(filters.Arg("event", "health_status")),
		})

		hadErrors := false

		for {
			if hadErrors || c.closed {
				break
			}

			select {
			case event := <-chMessage:
				parts := strings.Split(event.Status, ": ")
				if len(parts) == 2 {
					healthcheck.SetState(event.ID, healthcheck.NameToValue(parts[1]))
				}

			case err := <-chErr:
				if !c.closed {
					fmt.Println("Failed to watch for events from the engine:", err)
				}

				hadErrors = true
				cancel()
				break
			}
		}
	}
}
