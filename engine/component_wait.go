package engine

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types/container"
)

func (c *Component) WaitFor(exitChan chan<- ComponentExited) {
	if c.containerID == "" {
		exitChan <- ComponentExited{
			Component: c,
			Error:     errors.New("component not started"),
		}
		return
	}

	waitChan, errChan := c.client.api.ContainerWait(context.Background(), c.containerID, container.WaitConditionNotRunning)

	for {
		select {
		case exit := <-waitChan:
			if exit.Error != nil {
				exitChan <- ComponentExited{
					Component: c,
					Error:     errors.New(exit.Error.Message),
				}
			} else {
				exitChan <- ComponentExited{
					Component:  c,
					StatusCode: exit.StatusCode,
				}
			}

		case err := <-errChan:
			exitChan <- ComponentExited{
				Component: c,
				Error:     err,
			}
		}
	}
}
