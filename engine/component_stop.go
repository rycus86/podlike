package engine

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"strings"
	"time"
)

func (c *Component) Stop() error {
	fmt.Println("Stopping container:", c.Name)

	if c.containerID == "" {
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
	var (
		contextTimeout time.Duration
		stopTimeout    *time.Duration
	)
	if c.StopGracePeriod > 0 {
		contextTimeout = c.StopGracePeriod + 1*time.Second
		stopTimeout = &c.StopGracePeriod
	} else {
		contextTimeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	err := c.client.api.ContainerStop(ctx, c.containerID, stopTimeout)
	if err != nil {
		fmt.Println("Failed to stop the container:", err)
	}

	return err
}

func (c *Component) removeContainer() error {
	err := c.client.api.ContainerRemove(context.Background(), c.containerID, types.ContainerRemoveOptions{
		Force: true,
	})

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
		fmt.Sprintf("removal of container %s is already in progress", c.containerID),
	)
}
