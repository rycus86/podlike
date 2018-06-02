package component

import (
	"fmt"
	"io/ioutil"
)

func (c *Component) pullImage() error {
	fmt.Println("Pulling image:", c.Image)

	if reader, err := c.engine.PullImage(c.Image); err != nil {
		return err
	} else {
		defer reader.Close()

		ioutil.ReadAll(reader)

		return nil
	}
}
