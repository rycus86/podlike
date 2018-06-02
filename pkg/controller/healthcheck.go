package controller

import (
	"fmt"
	"github.com/rycus86/podlike/pkg/healthcheck"
	"strings"
)

func (c *Client) WatchHealthcheckEvents() {
	for {
		if c.closed {
			return
		}

		chMessage, chErr := c.engine.WatchHealthcheckEvents()

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
				break
			}
		}
	}
}
