package engine

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/rycus86/podlike/config"
	"time"
)

func (c *Component) Start(configuration *config.Configuration) error {
	fmt.Println("Starting component:", c.Name)

	containerID, err := c.createContainer(configuration)
	if err != nil {
		return err
	}

	// save the container ID for later
	c.containerID = containerID

	if err := c.copyFilesIfNecessary(); err != nil {
		return err
	}

	if err := c.startContainer(); err != nil {
		return err
	}

	if configuration.StreamLogs {
		go c.streamLogs()
	}

	fmt.Println("Component started:", c.Name)

	return nil
}

func (c *Component) startContainer() error {
	ctxStart, cancelStart := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelStart()

	return c.client.api.ContainerStart(ctxStart, c.containerID, types.ContainerStartOptions{})
}
