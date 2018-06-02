package component

import (
	"fmt"
	"github.com/rycus86/podlike/pkg/config"
	"github.com/rycus86/podlike/pkg/healthcheck"
)

func (c *Component) Start(configuration *config.Configuration) error {
	fmt.Println("Starting component:", c.Name)

	containerID, err := c.createContainer(configuration)
	if err != nil {
		return err
	}

	if err := c.readContainerJSON(containerID); err != nil {
		return err
	}

	if err := c.copyFilesIfNecessary(); err != nil {
		return err
	}

	if err := c.initHealthCheckingIfNecessary(); err != nil {
		return err
	}

	if err := c.startContainer(); err != nil {
		return err
	}

	healthcheck.MarkStarted(c.container.ID, c.Name)

	if configuration.StreamLogs {
		go c.streamLogs()
	}

	fmt.Println("Component started:", c.Name)

	return nil
}

func (c *Component) startContainer() error {
	return c.engine.StartContainer(c.container.ID)
}
