package component

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func (c *Component) Stop() error {
	fmt.Println("Stopping container:", c.Name)

	if c.container == nil {
		return errors.New("Container is not running for component: " + c.Name)
	}

	stopError := c.stopContainer()
	removeError := c.removeContainer()

	if removeError != nil {
		return removeError
	} else {
		return stopError
	}
}

func (c *Component) stopContainer() error {
	var stopTimeout *time.Duration

	if c.StopGracePeriod > 0 {
		stopTimeout = &c.StopGracePeriod
	}

	err := c.engine.StopContainer(c.container.ID, stopTimeout)
	if err != nil {
		fmt.Println("Failed to stop the container:", err)
	}

	return err
}

func (c *Component) removeContainer() error {
	err := c.engine.RemoveContainer(c.container.ID)

	if err != nil {
		if !c.isRemovalInProgressError(err) {
			fmt.Println("Failed to remove the container:", err)
		}
	}

	return err
}

func (c *Component) isRemovalInProgressError(err error) bool {
	// TODO this feels a bit hacky
	return strings.Contains(
		err.Error(),
		fmt.Sprintf("removal of container %s is already in progress", c.container.ID),
	)
}
