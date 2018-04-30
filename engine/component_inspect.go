package engine

import (
	"context"
	"fmt"
	"time"
)

func (c *Component) readContainerJSON(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ctr, err := c.client.api.ContainerInspect(ctx, containerID)
	if err != nil {
		fmt.Println("Could not determine whether", c.Name, "has healthchecks")
		return err
	}

	c.container = &ctr

	return nil
}
