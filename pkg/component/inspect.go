package component

import (
	"fmt"
)

func (c *Component) readContainerJSON(containerID string) error {
	ctr, err := c.engine.InspectContainer(containerID)
	if err != nil {
		fmt.Println("Could not determine whether", c.Name, "has healthchecks")
		return err
	}

	c.container = ctr

	return nil
}
