package component

import (
	"errors"
)

func (c *Component) WaitFor(exitChan chan<- ExitEvent) {
	if c.container == nil {
		exitChan <- ExitEvent{
			Component: c,
			Error:     errors.New("component not started"),
		}
		return
	}

	waitChan, errChan := c.engine.WaitContainer(c.container.ID)

	for {
		select {
		case exit := <-waitChan:
			if exit.Error != nil {
				exitChan <- ExitEvent{
					Component: c,
					Error:     errors.New(exit.Error.Message),
				}
			} else {
				exitChan <- ExitEvent{
					Component:  c,
					StatusCode: exit.StatusCode,
				}
			}

		case err := <-errChan:
			exitChan <- ExitEvent{
				Component: c,
				Error:     err,
			}
		}
	}
}
